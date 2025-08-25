package app

import (
	"context"
	"fmt"
	"iptv-gateway/internal/config"
	"iptv-gateway/internal/logging"
	"iptv-gateway/internal/metrics"
	"iptv-gateway/internal/urlgen"

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

	for subName := range cfg.Subscriptions {
		metrics.SubscriptionStreamsActive.WithLabelValues(subName).Set(0)
	}

	manager.subSemaphores = make(map[string]*semaphore.Weighted, len(cfg.Subscriptions))
	for subName, sub := range cfg.Subscriptions {
		if sub.Proxy.ConcurrentStreams > 0 {
			manager.subSemaphores[subName] = semaphore.NewWeighted(sub.Proxy.ConcurrentStreams)
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

func (m *Manager) GetGlobalSemaphore() *semaphore.Weighted {
	return m.semaphore
}

func (m *Manager) initClients() error {
	m.clients = make([]*Client, 0, len(m.config.Clients))

	for clientName, clientConf := range m.config.Clients {
		clientInstance, err := m.createClient(clientName, clientConf)
		if err != nil {
			return err
		}

		m.clients = append(m.clients, clientInstance)
		m.secretToClient[clientConf.Secret] = clientInstance

		logging.Debug(context.TODO(), "client initialized", "name", clientName)
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
		preset, found := m.config.Presets[presetName]
		if !found {
			return nil, fmt.Errorf("preset '%s' for client '%s' is not defined in config", presetName, clientName)
		}
		presets = append(presets, preset)
	}
	return presets, nil
}

func (m *Manager) addSubscriptionsToClient(clientInstance *Client, clientName string, clientConf config.Client) error {
	clientSubs := m.collectSubscriptions(clientInstance, clientConf)

	if len(clientSubs) == 0 {
		return fmt.Errorf("no subscriptions specified for %s", clientName)
	}

	for _, subName := range clientSubs {
		if err := m.addClientSubscription(clientInstance, clientName, clientConf, subName); err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) collectSubscriptions(clientInstance *Client, clientConf config.Client) []string {
	clientSubs := make([]string, 0, len(clientConf.Subscriptions)+len(clientInstance.presets)*2)
	clientSubs = append(clientSubs, clientConf.Subscriptions...)

	for _, preset := range clientInstance.presets {
		clientSubs = append(clientSubs, preset.Subscriptions...)
	}

	return clientSubs
}

func (m *Manager) addClientSubscription(clientInstance *Client, clientName string, clientConf config.Client, subName string) error {
	subConf, found := m.config.Subscriptions[subName]
	if !found {
		return fmt.Errorf("subscription '%s' for client '%s' is not defined in config", subName, clientName)
	}

	urlGen, err := m.createURLGenerator(clientConf.Secret, subName)
	if err != nil {
		return fmt.Errorf("failed to create URL generator: %w", err)
	}

	err = clientInstance.AddSubscription(
		subName, subConf, *urlGen,
		m.config.Rules, m.config.Proxy,
		m.subSemaphores[subName])
	if err != nil {
		return fmt.Errorf("failed to build subscription '%s' for client '%s': %w", subName, clientName, err)
	}

	return nil
}

func (m *Manager) createURLGenerator(clientSecret, subName string) (*urlgen.Generator, error) {
	baseURL := fmt.Sprintf("%s/%s", m.publicURLBase, clientSecret)
	secretKey := m.config.Secret + subName + clientSecret

	return urlgen.NewGenerator(baseURL, secretKey)
}
