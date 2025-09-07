package app

import (
	"context"
	"fmt"
	"iptv-gateway/internal/config"
	"iptv-gateway/internal/logging"
	"iptv-gateway/internal/metrics"
	"iptv-gateway/internal/urlgen"
	"time"

	"golang.org/x/sync/semaphore"
)

type Manager struct {
	config         *config.Config
	semaphore      *semaphore.Weighted
	subSemaphores  map[string]*semaphore.Weighted
	clients        []*Client
	secretToClient map[string]*Client
	publicURLBase  string
}

func NewManager(cfg *config.Config) (*Manager, error) {
	manager := &Manager{
		config:         cfg,
		secretToClient: make(map[string]*Client),
		publicURLBase:  cfg.PublicURL.String(),
	}

	if cfg.Proxy.Enabled != nil && *cfg.Proxy.Enabled && cfg.Proxy.ConcurrentStreams > 0 {
		manager.semaphore = semaphore.NewWeighted(cfg.Proxy.ConcurrentStreams)
	}

	for _, playlist := range cfg.Playlists {
		metrics.PlaylistStreamsActive.WithLabelValues(playlist.Name).Set(0)
	}

	manager.subSemaphores = make(map[string]*semaphore.Weighted, len(cfg.Playlists))
	for _, playlist := range cfg.Playlists {
		if playlist.Proxy.ConcurrentStreams > 0 {
			manager.subSemaphores[playlist.Name] = semaphore.NewWeighted(playlist.Proxy.ConcurrentStreams)
		}
	}

	if err := manager.initClients(); err != nil {
		return nil, err
	}

	return manager, nil
}

func (m *Manager) GetClient(secret string) *Client {
	return m.secretToClient[secret]
}

func (m *Manager) GlobalSemaphore() *semaphore.Weighted {
	return m.semaphore
}

func (m *Manager) initClients() error {
	m.clients = make([]*Client, 0, len(m.config.Clients))

	for _, clientConf := range m.config.Clients {
		clientInstance, err := m.createClient(clientConf.Name, clientConf)
		if err != nil {
			return err
		}

		m.clients = append(m.clients, clientInstance)
		m.secretToClient[clientConf.Secret] = clientInstance

		logging.Debug(context.TODO(), "client initialized", "name", clientConf.Name)
	}
	return nil
}

func (m *Manager) createClient(clientName string, clientConf config.Client) (*Client, error) {
	presets, err := m.resolvePresets(clientName, clientConf.Preset)
	if err != nil {
		return nil, err
	}

	clientInstance, err := NewClient(clientName, clientConf, presets, m.publicURLBase)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize client %s: %w", clientName, err)
	}

	if err := m.addSubscriptionsToClient(clientInstance, clientName, clientConf); err != nil {
		return nil, fmt.Errorf("failed to add subscriptions for client %s: %w", clientName, err)
	}

	return clientInstance, nil
}

func (m *Manager) resolvePresets(clientName string, presetNames []string) ([]config.Preset, error) {
	if len(presetNames) == 0 {
		return nil, nil
	}

	presets := make([]config.Preset, 0, len(presetNames))
	for _, presetName := range presetNames {
		var preset config.Preset
		found := false
		for _, p := range m.config.Presets {
			if p.Name == presetName {
				preset = p
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("preset '%s' for client '%s' is not defined in config", presetName, clientName)
		}
		presets = append(presets, preset)
	}
	return presets, nil
}

func (m *Manager) addSubscriptionsToClient(clientInstance *Client, clientName string, clientConf config.Client) error {
	clientPlaylistNames := m.collectPlaylists(clientInstance, clientConf)
	for _, playlistName := range clientPlaylistNames {
		if err := m.addPlaylistSubscription(clientInstance, clientName, clientConf, playlistName); err != nil {
			return err
		}
	}

	clientEPGNames := m.collectEPGs(clientInstance, clientConf)
	for _, epgName := range clientEPGNames {
		if err := m.addEPGSubscription(clientInstance, clientName, clientConf, epgName); err != nil {
			return err
		}
	}

	if len(clientPlaylistNames) == 0 && len(clientEPGNames) == 0 {
		return fmt.Errorf("no playlists or EPGs specified for %s", clientName)
	}

	return nil
}

func (m *Manager) collectPlaylists(clientInstance *Client, clientConf config.Client) []string {
	return collectNames(clientConf, clientInstance,
		func(c config.Client) []string { return c.Playlists },
		func(p config.Preset) []string { return p.Playlists },
	)
}

func (m *Manager) collectEPGs(clientInstance *Client, clientConf config.Client) []string {
	return collectNames(clientConf, clientInstance,
		func(c config.Client) []string { return c.EPGs },
		func(p config.Preset) []string { return p.EPGs },
	)
}

func (m *Manager) addPlaylistSubscription(
	clientInstance *Client, clientName string, clientConf config.Client, playlistName string) error {
	var playlistConf config.Playlist

	found := false
	for _, playlist := range m.config.Playlists {
		if playlist.Name == playlistName {
			playlistConf = playlist
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("playlist '%s' for client '%s' is not defined in config", playlistName, clientName)
	}

	urlGen, err := m.createURLGenerator(clientConf.Secret, playlistName)
	if err != nil {
		return fmt.Errorf("failed to create URL generator: %w", err)
	}

	err = clientInstance.BuildPlaylistSubscription(
		playlistConf, *urlGen,
		m.config.ChannelRules, m.config.PlaylistRules,
		m.config.Proxy,
		m.subSemaphores[playlistName],
		m.config.Conditions)
	if err != nil {
		return fmt.Errorf("failed to build playlist subscription '%s' for client '%s': %w", playlistName, clientName, err)
	}

	return nil
}

func (m *Manager) addEPGSubscription(
	clientInstance *Client, clientName string, clientConf config.Client, epgName string) error {
	var epgConf config.EPG

	found := false
	for _, epg := range m.config.EPGs {
		if epg.Name == epgName {
			epgConf = epg
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("EPG '%s' for client '%s' is not defined in config", epgName, clientName)
	}

	urlGen, err := m.createURLGenerator(clientConf.Secret, epgName)
	if err != nil {
		return fmt.Errorf("failed to create URL generator: %w", err)
	}

	err = clientInstance.BuildEPGSubscription(epgConf, *urlGen, m.config.Proxy)
	if err != nil {
		return fmt.Errorf("failed to build EPG subscription '%s' for client '%s': %w", epgName, clientName, err)
	}

	return nil
}

func (m *Manager) createURLGenerator(clientSecret, srcName string) (*urlgen.Generator, error) {
	baseURL := fmt.Sprintf("%s/%s", m.publicURLBase, clientSecret)
	secretKey := m.config.URLGenerator.Secret + srcName + clientSecret

	return urlgen.NewGenerator(
		baseURL, secretKey,
		time.Duration(m.config.URLGenerator.StreamTTL),
		time.Duration(m.config.URLGenerator.FileTTL),
	)
}

func collectNames(clientConf config.Client, clientInstance *Client,
	getClientNames func(config.Client) []string,
	getPresetNames func(config.Preset) []string) []string {

	nameSet := make(map[string]bool)

	for _, name := range getClientNames(clientConf) {
		nameSet[name] = true
	}

	for _, preset := range clientInstance.presets {
		for _, name := range getPresetNames(preset) {
			nameSet[name] = true
		}
	}

	names := make([]string, 0, len(nameSet))
	for name := range nameSet {
		names = append(names, name)
	}

	return names
}
