package connectors

import (
	"fmt"
	"sync"
)

// ConnectorFactory is a function that creates a new CRMConnector instance
type ConnectorFactory func(config ConnectorConfig) (CRMConnector, error)

// Registry manages available CRM connector implementations
type Registry struct {
	mu        sync.RWMutex
	factories map[string]ConnectorFactory
}

var defaultRegistry = &Registry{
	factories: make(map[string]ConnectorFactory),
}

// Register adds a connector factory to the default registry
func Register(platformSlug string, factory ConnectorFactory) {
	defaultRegistry.mu.Lock()
	defer defaultRegistry.mu.Unlock()
	defaultRegistry.factories[platformSlug] = factory
}

// NewConnector creates a new CRM connector by platform slug
func NewConnector(platformSlug string, config ConnectorConfig) (CRMConnector, error) {
	defaultRegistry.mu.RLock()
	defer defaultRegistry.mu.RUnlock()

	factory, ok := defaultRegistry.factories[platformSlug]
	if !ok {
		return nil, fmt.Errorf("no connector registered for platform: %s", platformSlug)
	}

	return factory(config)
}

// ListRegistered returns all registered platform slugs
func ListRegistered() []string {
	defaultRegistry.mu.RLock()
	defer defaultRegistry.mu.RUnlock()

	slugs := make([]string, 0, len(defaultRegistry.factories))
	for slug := range defaultRegistry.factories {
		slugs = append(slugs, slug)
	}
	return slugs
}

// IsRegistered checks if a connector is registered for the given platform slug
func IsRegistered(platformSlug string) bool {
	defaultRegistry.mu.RLock()
	defer defaultRegistry.mu.RUnlock()

	_, ok := defaultRegistry.factories[platformSlug]
	return ok
}
