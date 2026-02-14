package translate

import (
	"context"
	"log"

	"github.com/myfusionhelper/api/internal/connectors"
)

// TranslatingConnector wraps any CRMConnector and transparently translates
// field names, resolves custom field labels to IDs, resolves tag names to IDs,
// and normalizes data formats on read/write.
//
// It implements the CRMConnector interface so it can be used as a drop-in
// replacement anywhere a connector is expected.
type TranslatingConnector struct {
	inner        connectors.CRMConnector
	fieldMapper  *FieldMapper
	customFields *CustomFieldResolver
	tagResolver  *TagResolver
	normalizer   *DataNormalizer
}

// NewTranslatingConnector wraps a raw CRMConnector with the translation layer.
// The platform slug is determined from the inner connector's metadata.
func NewTranslatingConnector(inner connectors.CRMConnector) *TranslatingConnector {
	platformSlug := inner.GetMetadata().PlatformSlug

	return &TranslatingConnector{
		inner:        inner,
		fieldMapper:  NewFieldMapper(platformSlug),
		customFields: NewCustomFieldResolver(inner),
		tagResolver:  NewTagResolver(inner),
		normalizer:   NewDataNormalizer(platformSlug),
	}
}

// resolveFieldKey translates a user-facing field key to the CRM-specific key.
// It first checks static standard field mappings, then falls back to
// custom field label resolution.
func (t *TranslatingConnector) resolveFieldKey(ctx context.Context, fieldKey string) string {
	// Step 1: Check if it's a standard field name
	resolved := t.fieldMapper.Resolve(fieldKey)
	if resolved != fieldKey {
		return resolved
	}

	// Step 2: If the key is already a CRM-native standard key, pass through
	if t.fieldMapper.IsCRMNativeKey(fieldKey) {
		return fieldKey
	}

	// Step 3: Try custom field label/key resolution
	customID, err := t.customFields.ResolveLabel(ctx, fieldKey)
	if err != nil {
		log.Printf("translate: warning: failed to resolve custom field %q: %v", fieldKey, err)
		return fieldKey
	}
	if customID != "" {
		return customID
	}

	// Not resolved — return as-is (may be a raw CRM-specific ID)
	return fieldKey
}

// ========== CONTACTS ==========

func (t *TranslatingConnector) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return t.inner.GetContacts(ctx, opts)
}

func (t *TranslatingConnector) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	return t.inner.GetContact(ctx, contactID)
}

func (t *TranslatingConnector) CreateContact(ctx context.Context, contact connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return t.inner.CreateContact(ctx, contact)
}

func (t *TranslatingConnector) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return t.inner.UpdateContact(ctx, contactID, updates)
}

func (t *TranslatingConnector) DeleteContact(ctx context.Context, contactID string) error {
	return t.inner.DeleteContact(ctx, contactID)
}

// ========== FIELD ACCESS (INTERCEPTED) ==========

func (t *TranslatingConnector) GetContactFieldValue(ctx context.Context, contactID string, fieldKey string) (interface{}, error) {
	// Translate field key
	resolvedKey := t.resolveFieldKey(ctx, fieldKey)

	// Call inner connector with resolved key
	value, err := t.inner.GetContactFieldValue(ctx, contactID, resolvedKey)
	if err != nil {
		return nil, err
	}

	// Normalize output data format
	fieldType := t.customFields.GetFieldType(ctx, resolvedKey)
	if fieldType != "" {
		value = t.normalizer.NormalizeRead(value, fieldType)
	}

	return value, nil
}

func (t *TranslatingConnector) SetContactFieldValue(ctx context.Context, contactID string, fieldKey string, value interface{}) error {
	// Translate field key
	resolvedKey := t.resolveFieldKey(ctx, fieldKey)

	// Convert value to CRM-specific format
	fieldType := t.customFields.GetFieldType(ctx, resolvedKey)
	if fieldType != "" {
		value = t.normalizer.NormalizeWrite(value, fieldType)
	}

	// Call inner connector with resolved key and converted value
	return t.inner.SetContactFieldValue(ctx, contactID, resolvedKey, value)
}

// ========== TAGS (INTERCEPTED) ==========

func (t *TranslatingConnector) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return t.inner.GetTags(ctx)
}

func (t *TranslatingConnector) ApplyTag(ctx context.Context, contactID string, tagID string) error {
	resolvedID, err := t.tagResolver.Resolve(ctx, tagID)
	if err != nil {
		return err
	}
	return t.inner.ApplyTag(ctx, contactID, resolvedID)
}

func (t *TranslatingConnector) RemoveTag(ctx context.Context, contactID string, tagID string) error {
	resolvedID, err := t.tagResolver.Resolve(ctx, tagID)
	if err != nil {
		return err
	}
	return t.inner.RemoveTag(ctx, contactID, resolvedID)
}

// ========== CUSTOM FIELDS ==========

func (t *TranslatingConnector) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return t.inner.GetCustomFields(ctx)
}

// ========== AUTOMATIONS ==========

func (t *TranslatingConnector) TriggerAutomation(ctx context.Context, contactID string, automationID string) error {
	return t.inner.TriggerAutomation(ctx, contactID, automationID)
}

func (t *TranslatingConnector) AchieveGoal(ctx context.Context, contactID string, goalName string, integration string) error {
	return t.inner.AchieveGoal(ctx, contactID, goalName, integration)
}

// ========== MARKETING / OPT-IN ==========

func (t *TranslatingConnector) SetOptInStatus(ctx context.Context, contactID string, optIn bool, reason string) error {
	return t.inner.SetOptInStatus(ctx, contactID, optIn, reason)
}

// ========== HEALTH & METADATA ==========

func (t *TranslatingConnector) TestConnection(ctx context.Context) error {
	return t.inner.TestConnection(ctx)
}

func (t *TranslatingConnector) GetMetadata() connectors.ConnectorMetadata {
	return t.inner.GetMetadata()
}

func (t *TranslatingConnector) GetCapabilities() []connectors.Capability {
	return t.inner.GetCapabilities()
}
