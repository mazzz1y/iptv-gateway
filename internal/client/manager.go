package client

import (
	"context"
	"fmt"
	"iptv-gateway/internal/config"
	"iptv-gateway/internal/logging"
	"iptv-gateway/internal/urlgen"

	"golang.org/x/sync/semaphore"
)

type Manager struct {
	config        *config.Config
	semaphore     *semaphore.Weighted
	subSemaphores map[string]*semaphore.Weighted
	clients       []*Client
}

func NewManager(cfg *config.Config) (*Manager, error) {
	var sem *semaphore.Weighted
	if cfg.Proxy.Enabled != nil && *cfg.Proxy.Enabled && cfg.Proxy.ConcurrentStreams > 0 {
		sem = semaphore.NewWeighted(cfg.Proxy.ConcurrentStreams)
	}

	subSemaphores := make(map[string]*semaphore.Weighted)
	for subName, sub := range cfg.Subscriptions {
		if sub.Proxy.ConcurrentStreams > 0 {
			subSemaphores[subName] = semaphore.NewWeighted(sub.Proxy.ConcurrentStreams)
		}
	}

	manager := &Manager{
		config:        cfg,
		semaphore:     sem,
		subSemaphores: subSemaphores,
	}

	if err := manager.initClients(); err != nil {
		return nil, err
	}

	return manager, nil
}

func (m *Manager) GetClient(secret string) *Client {
	for _, c := range m.clients {
		if c.IsCorrectSecret(secret) {
			return c
		}
	}
	return nil
}

func (m *Manager) GetSemaphore() *semaphore.Weighted {
	return m.semaphore
}

func (m *Manager) initClients() error {
	for clientName, clientConf := range m.config.Clients {

		presets := make([]config.Preset, 0, len(clientConf.Preset))
		for _, presetName := range clientConf.Preset {
			preset, found := m.config.Presets[presetName]
			if !found {
				return fmt.Errorf("preset '%s' for client '%s' is not defined in config", presetName, clientName)
			}
			presets = append(presets, preset)
		}

		clientInstance, err := NewClient(clientName, clientConf, presets, m.config.PublicURL.String())
		if err != nil {
			return fmt.Errorf("failed to initialize client %s: %w", clientName, err)
		}

		if err := m.addSubscriptionsToClient(clientInstance, clientName, clientConf); err != nil {
			return fmt.Errorf("failed to add subscriptions for client %s: %w", clientName, err)
		}

		m.clients = append(m.clients, clientInstance)
		logging.Debug(context.TODO(), "client initialized", "name", clientName)
	}
	return nil
}

func (m *Manager) addSubscriptionsToClient(clientInstance *Client, clientName string, clientConf config.Client) error {
	if len(clientConf.Subscriptions) == 0 {
		return fmt.Errorf("no subscriptions specified for %s", clientName)
	}

	for _, subName := range clientConf.Subscriptions {
		subConf, found := m.config.Subscriptions[subName]
		if !found {
			return fmt.Errorf("subscription '%s' for client '%s' is not defined in config", subName, clientName)
		}

		urlGen, err := m.createURLGenerator(clientConf.Secret, subName)
		if err != nil {
			return fmt.Errorf("failed to create URL generator: %w", err)
		}

		err = clientInstance.AddSubscription(
			subName, subConf, urlGen,
			m.config.Rules, m.config.Proxy,
			m.subSemaphores[subName])
		if err != nil {
			return fmt.Errorf("failed to build subscription '%s' for client '%s': %w", subName, clientName, err)
		}
	}

	return nil
}

func (m *Manager) createURLGenerator(clientSecret, subName string) (*urlgen.Generator, error) {
	baseURL := fmt.Sprintf("%s/%s", m.config.PublicURL.String(), clientSecret)
	secretKey := m.config.Secret + subName + clientSecret

	return urlgen.NewGenerator(baseURL, secretKey)
}
