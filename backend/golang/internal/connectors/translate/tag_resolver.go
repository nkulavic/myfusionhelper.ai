package translate

import (
	"context"
	"strings"
	"sync"

	"github.com/myfusionhelper/api/internal/connectors"
)

// TagResolver lazily loads tags from the CRM and resolves tag names to IDs.
type TagResolver struct {
	inner    connectors.CRMConnector
	mu       sync.RWMutex
	loaded   bool
	nameToID map[string]string // normalized_name -> tag_id
	idExists map[string]bool   // quick check if a value is already an ID
}

// NewTagResolver creates a resolver backed by the given connector.
func NewTagResolver(inner connectors.CRMConnector) *TagResolver {
	return &TagResolver{
		inner:    inner,
		nameToID: make(map[string]string),
		idExists: make(map[string]bool),
	}
}

// ensureLoaded lazily loads tags on first use.
func (r *TagResolver) ensureLoaded(ctx context.Context) error {
	r.mu.RLock()
	if r.loaded {
		r.mu.RUnlock()
		return nil
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.loaded {
		return nil
	}

	tags, err := r.inner.GetTags(ctx)
	if err != nil {
		return err
	}

	for _, t := range tags {
		normalizedName := strings.ToLower(strings.TrimSpace(t.Name))
		r.nameToID[normalizedName] = t.ID
		r.idExists[t.ID] = true
	}

	r.loaded = true
	return nil
}

// Resolve translates a tag name or ID to the CRM-specific tag ID.
// If the input is already a valid tag ID, it is returned as-is.
// If the input is a tag name, it is resolved to the corresponding ID.
// If resolution fails or the tag is not found, the original value is returned
// so the inner connector can handle it (some CRMs accept tag names directly).
func (r *TagResolver) Resolve(ctx context.Context, tagNameOrID string) (string, error) {
	if err := r.ensureLoaded(ctx); err != nil {
		// If we cannot load tags, pass through the original value.
		// The inner connector will handle any errors.
		return tagNameOrID, nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	// If it's already a known tag ID, return as-is
	if r.idExists[tagNameOrID] {
		return tagNameOrID, nil
	}

	// Try name lookup (case-insensitive)
	normalized := strings.ToLower(strings.TrimSpace(tagNameOrID))
	if id, ok := r.nameToID[normalized]; ok {
		return id, nil
	}

	// Not found in name map and not a known ID.
	// Return as-is — might be a valid ID not loaded (pagination limit),
	// or the CRM accepts tag names directly (e.g., GoHighLevel).
	return tagNameOrID, nil
}
