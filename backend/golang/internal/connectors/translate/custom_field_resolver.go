package translate

import (
	"context"
	"strings"
	"sync"

	"github.com/myfusionhelper/api/internal/connectors"
)

// CustomFieldResolver lazily loads custom field definitions from the CRM
// and provides label-to-ID and key-to-ID resolution with caching.
type CustomFieldResolver struct {
	inner      connectors.CRMConnector
	mu         sync.RWMutex
	loaded     bool
	labelToID  map[string]string // normalized_label -> field_id
	keyToID    map[string]string // field_key -> field_id
	fieldTypes map[string]string // field_id_or_key -> field_type
	fields     []connectors.CustomField
}

// NewCustomFieldResolver creates a resolver backed by the given connector.
func NewCustomFieldResolver(inner connectors.CRMConnector) *CustomFieldResolver {
	return &CustomFieldResolver{
		inner:      inner,
		labelToID:  make(map[string]string),
		keyToID:    make(map[string]string),
		fieldTypes: make(map[string]string),
	}
}

// ensureLoaded lazily loads custom fields on first use.
// Uses double-check locking for thread safety.
func (r *CustomFieldResolver) ensureLoaded(ctx context.Context) error {
	r.mu.RLock()
	if r.loaded {
		r.mu.RUnlock()
		return nil
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check after acquiring write lock
	if r.loaded {
		return nil
	}

	fields, err := r.inner.GetCustomFields(ctx)
	if err != nil {
		return err
	}

	r.fields = fields
	for _, f := range fields {
		normalizedLabel := normalizeLabel(f.Label)
		r.labelToID[normalizedLabel] = f.ID

		if f.Key != "" {
			r.keyToID[f.Key] = f.ID
			// Also index by normalized key for flexible matching
			r.keyToID[normalizeLabel(f.Key)] = f.ID
		}

		r.fieldTypes[f.ID] = f.FieldType
		if f.Key != "" {
			r.fieldTypes[f.Key] = f.FieldType
		}
	}

	r.loaded = true
	return nil
}

// ResolveLabel translates a human-readable custom field label or key to a CRM-specific ID.
// Returns empty string if no match found (the input may already be a raw CRM-specific ID).
func (r *CustomFieldResolver) ResolveLabel(ctx context.Context, label string) (string, error) {
	if err := r.ensureLoaded(ctx); err != nil {
		return "", err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	// Try exact normalized label match
	normalized := normalizeLabel(label)
	if id, ok := r.labelToID[normalized]; ok {
		return id, nil
	}

	// Try key match (e.g., "lead_score" matches a Key field)
	if id, ok := r.keyToID[label]; ok {
		return id, nil
	}

	// Try normalized key match
	if id, ok := r.keyToID[normalized]; ok {
		return id, nil
	}

	// Not found — the input is probably already a raw CRM-specific ID
	return "", nil
}

// GetFieldType returns the CRM field type for a resolved field ID or key.
// Returns empty string for standard fields or unknown custom fields.
func (r *CustomFieldResolver) GetFieldType(ctx context.Context, fieldIDOrKey string) string {
	// Don't trigger a load just for field type — only return if already loaded
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.loaded {
		return ""
	}

	return r.fieldTypes[fieldIDOrKey]
}

// normalizeLabel lowercases and trims a label for case-insensitive matching.
func normalizeLabel(label string) string {
	return strings.ToLower(strings.TrimSpace(label))
}
