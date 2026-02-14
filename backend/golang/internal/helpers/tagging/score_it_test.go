package tagging

import (
	"context"
	"fmt"
	"testing"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

type mockConnectorForScoreIt struct {
	contact         *connectors.NormalizedContact
	getContactError error
	setFieldError   error
	fieldsSet       map[string]interface{}
}

func (m *mockConnectorForScoreIt) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	if m.getContactError != nil {
		return nil, m.getContactError
	}
	return m.contact, nil
}

func (m *mockConnectorForScoreIt) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	if m.setFieldError != nil {
		return m.setFieldError
	}
	if m.fieldsSet == nil {
		m.fieldsSet = make(map[string]interface{})
	}
	m.fieldsSet[fieldKey] = value
	return nil
}

func (m *mockConnectorForScoreIt) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForScoreIt) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForScoreIt) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForScoreIt) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForScoreIt) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForScoreIt) ApplyTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForScoreIt) RemoveTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForScoreIt) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForScoreIt) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockConnectorForScoreIt) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForScoreIt) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForScoreIt) TestConnection(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}

func (m *mockConnectorForScoreIt) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{PlatformSlug: "mock", PlatformName: "Mock"}
}

func (m *mockConnectorForScoreIt) GetCapabilities() []connectors.Capability {
	return []connectors.Capability{}
}

func TestScoreIt_GetMetadata(t *testing.T) {
	helper := &ScoreIt{}
	if helper.GetName() != "Score It" {
		t.Errorf("Expected name 'Score It', got '%s'", helper.GetName())
	}
	if helper.GetType() != "score_it" {
		t.Errorf("Expected type 'score_it', got '%s'", helper.GetType())
	}
	if helper.GetCategory() != "tagging" {
		t.Errorf("Expected category 'tagging', got '%s'", helper.GetCategory())
	}
}

func TestScoreIt_ValidateConfig_Success(t *testing.T) {
	helper := &ScoreIt{}
	err := helper.ValidateConfig(map[string]interface{}{
		"rules": []interface{}{
			map[string]interface{}{"tag_id": "tag1", "has_tag": true, "points": 10.0},
		},
		"target_field": "score",
	})
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestScoreIt_ValidateConfig_MissingRules(t *testing.T) {
	helper := &ScoreIt{}
	err := helper.ValidateConfig(map[string]interface{}{"target_field": "score"})
	if err == nil {
		t.Error("Expected error for missing rules")
	}
}

func TestScoreIt_ValidateConfig_EmptyRules(t *testing.T) {
	helper := &ScoreIt{}
	err := helper.ValidateConfig(map[string]interface{}{
		"rules":        []interface{}{},
		"target_field": "score",
	})
	if err == nil {
		t.Error("Expected error for empty rules array")
	}
}

func TestScoreIt_ValidateConfig_MissingTargetField(t *testing.T) {
	helper := &ScoreIt{}
	err := helper.ValidateConfig(map[string]interface{}{
		"rules": []interface{}{
			map[string]interface{}{"tag_id": "tag1"},
		},
	})
	if err == nil {
		t.Error("Expected error for missing target_field")
	}
}

func TestScoreIt_Execute_BasicScoring(t *testing.T) {
	helper := &ScoreIt{}
	mock := &mockConnectorForScoreIt{
		contact: &connectors.NormalizedContact{
			ID: "123",
			Tags: []connectors.TagRef{
				{ID: "tag1", Name: "VIP"},
				{ID: "tag2", Name: "Premium"},
			},
		},
	}

	output, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"rules": []interface{}{
				map[string]interface{}{"tag_id": "tag1", "has_tag": true, "points": 10.0},
				map[string]interface{}{"tag_id": "tag2", "has_tag": true, "points": 5.0},
			},
			"target_field": "score",
		},
		Connector: mock,
	})

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !output.Success {
		t.Error("Expected success")
	}
	if mock.fieldsSet["score"] != "15" {
		t.Errorf("Expected score to be '15', got: %v", mock.fieldsSet["score"])
	}
}

func TestScoreIt_Execute_HasTagFalse(t *testing.T) {
	helper := &ScoreIt{}
	mock := &mockConnectorForScoreIt{
		contact: &connectors.NormalizedContact{
			ID:   "123",
			Tags: []connectors.TagRef{{ID: "tag1"}},
		},
	}

	_, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"rules": []interface{}{
				map[string]interface{}{"tag_id": "tag1", "has_tag": false, "points": 10.0},
			},
			"target_field": "score",
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if mock.fieldsSet["score"] != "0" {
		t.Errorf("Expected score '0' (has_tag=false but contact has tag), got: %v", mock.fieldsSet["score"])
	}
}

func TestScoreIt_Execute_NegativePoints(t *testing.T) {
	helper := &ScoreIt{}
	mock := &mockConnectorForScoreIt{
		contact: &connectors.NormalizedContact{
			ID:   "123",
			Tags: []connectors.TagRef{{ID: "tag1"}},
		},
	}

	_, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"rules": []interface{}{
				map[string]interface{}{"tag_id": "tag1", "has_tag": true, "points": -5.0},
			},
			"target_field": "score",
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if mock.fieldsSet["score"] != "-5" {
		t.Errorf("Expected score '-5', got: %v", mock.fieldsSet["score"])
	}
}

func TestScoreIt_Execute_MixedRules(t *testing.T) {
	helper := &ScoreIt{}
	mock := &mockConnectorForScoreIt{
		contact: &connectors.NormalizedContact{
			ID:   "123",
			Tags: []connectors.TagRef{{ID: "tag1"}},
		},
	}

	output, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"rules": []interface{}{
				map[string]interface{}{"tag_id": "tag1", "has_tag": true, "points": 10.0},
				map[string]interface{}{"tag_id": "tag2", "has_tag": false, "points": 5.0},
				map[string]interface{}{"tag_id": "tag3", "has_tag": true, "points": 3.0},
			},
			"target_field": "score",
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if mock.fieldsSet["score"] != "15" {
		t.Errorf("Expected score '15' (tag1: +10, tag2: +5, tag3: 0), got: %v", mock.fieldsSet["score"])
	}
	if !output.Success {
		t.Error("Expected success")
	}
}

func TestScoreIt_Execute_NoMatchingRules(t *testing.T) {
	helper := &ScoreIt{}
	mock := &mockConnectorForScoreIt{
		contact: &connectors.NormalizedContact{
			ID:   "123",
			Tags: []connectors.TagRef{},
		},
	}

	_, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"rules": []interface{}{
				map[string]interface{}{"tag_id": "tag1", "has_tag": true, "points": 10.0},
			},
			"target_field": "score",
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if mock.fieldsSet["score"] != "0" {
		t.Errorf("Expected score '0', got: %v", mock.fieldsSet["score"])
	}
}

func TestScoreIt_Execute_DefaultValues(t *testing.T) {
	helper := &ScoreIt{}
	mock := &mockConnectorForScoreIt{
		contact: &connectors.NormalizedContact{
			ID:   "123",
			Tags: []connectors.TagRef{{ID: "tag1"}},
		},
	}

	_, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"rules": []interface{}{
				map[string]interface{}{"tag_id": "tag1"},
			},
			"target_field": "score",
		},
		Connector: mock,
	})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if mock.fieldsSet["score"] != "0" {
		t.Errorf("Expected score '0' (default has_tag=true, points=0), got: %v", mock.fieldsSet["score"])
	}
}

func TestScoreIt_Execute_GetContactError(t *testing.T) {
	helper := &ScoreIt{}
	mock := &mockConnectorForScoreIt{
		getContactError: fmt.Errorf("contact error"),
	}

	_, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"rules":        []interface{}{map[string]interface{}{"tag_id": "tag1"}},
			"target_field": "score",
		},
		Connector: mock,
	})

	if err == nil {
		t.Error("Expected error for GetContact failure")
	}
}

func TestScoreIt_Execute_SetFieldError(t *testing.T) {
	helper := &ScoreIt{}
	mock := &mockConnectorForScoreIt{
		contact:       &connectors.NormalizedContact{ID: "123", Tags: []connectors.TagRef{}},
		setFieldError: fmt.Errorf("field error"),
	}

	_, err := helper.Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"rules":        []interface{}{map[string]interface{}{"tag_id": "tag1"}},
			"target_field": "score",
		},
		Connector: mock,
	})

	if err == nil {
		t.Error("Expected error for SetContactFieldValue failure")
	}
}
