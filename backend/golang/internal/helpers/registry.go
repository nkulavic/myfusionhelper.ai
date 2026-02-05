package helpers

import (
	"fmt"
	"sync"
)

// HelperFactory creates a new Helper instance
type HelperFactory func() Helper

// Registry manages available helper implementations
type Registry struct {
	mu        sync.RWMutex
	factories map[string]HelperFactory
}

var defaultRegistry = &Registry{
	factories: make(map[string]HelperFactory),
}

// Register adds a helper factory to the default registry
func Register(helperType string, factory HelperFactory) {
	defaultRegistry.mu.Lock()
	defer defaultRegistry.mu.Unlock()
	defaultRegistry.factories[helperType] = factory
}

// NewHelper creates a new helper instance by type
func NewHelper(helperType string) (Helper, error) {
	defaultRegistry.mu.RLock()
	defer defaultRegistry.mu.RUnlock()

	factory, ok := defaultRegistry.factories[helperType]
	if !ok {
		return nil, fmt.Errorf("unknown helper type: %s", helperType)
	}

	return factory(), nil
}

// ListHelperTypes returns all registered helper types
func ListHelperTypes() []string {
	defaultRegistry.mu.RLock()
	defer defaultRegistry.mu.RUnlock()

	types := make([]string, 0, len(defaultRegistry.factories))
	for t := range defaultRegistry.factories {
		types = append(types, t)
	}
	return types
}

// IsRegistered checks if a helper type is registered
func IsRegistered(helperType string) bool {
	defaultRegistry.mu.RLock()
	defer defaultRegistry.mu.RUnlock()

	_, ok := defaultRegistry.factories[helperType]
	return ok
}

// HelperInfo provides metadata about a registered helper
type HelperInfo struct {
	Type         string                 `json:"type"`
	Name         string                 `json:"name"`
	Category     string                 `json:"category"`
	Description  string                 `json:"description"`
	RequiresCRM  bool                   `json:"requires_crm"`
	SupportedCRMs []string             `json:"supported_crms"`
	ConfigSchema map[string]interface{} `json:"config_schema"`
}

// ListHelperInfo returns metadata about all registered helpers
func ListHelperInfo() []HelperInfo {
	defaultRegistry.mu.RLock()
	defer defaultRegistry.mu.RUnlock()

	infos := make([]HelperInfo, 0, len(defaultRegistry.factories))
	for helperType, factory := range defaultRegistry.factories {
		h := factory()
		infos = append(infos, HelperInfo{
			Type:         helperType,
			Name:         h.GetName(),
			Category:     h.GetCategory(),
			Description:  h.GetDescription(),
			RequiresCRM:  h.RequiresCRM(),
			SupportedCRMs: h.SupportedCRMs(),
			ConfigSchema: h.GetConfigSchema(),
		})
	}
	return infos
}
