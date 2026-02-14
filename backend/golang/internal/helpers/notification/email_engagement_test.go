package notification

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

type mockConnectorForEmailEngagement struct {
	fieldValues   map[string]interface{}
	getFieldError map[string]error
	setFieldError map[string]error
	tagsApplied   []string
	tagsRemoved   []string
	applyTagError map[string]error
}

func (m *mockConnectorForEmailEngagement) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	if m.getFieldError != nil {
		if err, ok := m.getFieldError[fieldKey]; ok {
			return nil, err
		}
	}
	if m.fieldValues != nil {
		return m.fieldValues[fieldKey], nil
	}
	return nil, nil
}

func (m *mockConnectorForEmailEngagement) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	if m.setFieldError != nil {
		if err, ok := m.setFieldError[fieldKey]; ok {
			return err
		}
	}
	if m.fieldValues == nil {
		m.fieldValues = make(map[string]interface{})
	}
	m.fieldValues[fieldKey] = value
	return nil
}

func (m *mockConnectorForEmailEngagement) ApplyTag(ctx context.Context, contactID, tagID string) error {
	if m.applyTagError != nil {
		if err, ok := m.applyTagError[tagID]; ok {
			return err
		}
	}
	if m.tagsApplied == nil {
		m.tagsApplied = make([]string, 0)
	}
	m.tagsApplied = append(m.tagsApplied, tagID)
	return nil
}

func (m *mockConnectorForEmailEngagement) RemoveTag(ctx context.Context, contactID, tagID string) error {
	if m.tagsRemoved == nil {
		m.tagsRemoved = make([]string, 0)
	}
	m.tagsRemoved = append(m.tagsRemoved, tagID)
	return nil
}

func (m *mockConnectorForEmailEngagement) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForEmailEngagement) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForEmailEngagement) CreateContact(ctx context.Context, contact connectors.CreateContactInput) (*connectors.NormalizedContact, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForEmailEngagement) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForEmailEngagement) DeleteContact(ctx context.Context, contactID string) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForEmailEngagement) GetTags(ctx context.Context) ([]connectors.Tag, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForEmailEngagement) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) { return nil, fmt.Errorf("not implemented") }
func (m *mockConnectorForEmailEngagement) TriggerAutomation(ctx context.Context, contactID, automationID string) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForEmailEngagement) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForEmailEngagement) TestConnection(ctx context.Context) error { return fmt.Errorf("not implemented") }
func (m *mockConnectorForEmailEngagement) GetMetadata() connectors.ConnectorMetadata { return connectors.ConnectorMetadata{} }
func (m *mockConnectorForEmailEngagement) GetCapabilities() []connectors.Capability { return nil }

func TestEmailEngagement_GetMetadata(t *testing.T) {
	h := &EmailEngagement{}
	if h.GetName() != "Email Engagement" { t.Error("wrong name") }
	if h.GetType() != "email_engagement" { t.Error("wrong type") }
	if h.GetCategory() != "notification" { t.Error("wrong category") }
	if !h.RequiresCRM() { t.Error("should require CRM") }
	if h.SupportedCRMs()[0] != "keap" { t.Error("should support keap") }
}

func TestEmailEngagement_ValidateConfig_InvalidEngagementType(t *testing.T) {
	err := (&EmailEngagement{}).ValidateConfig(map[string]interface{}{
		"engagement_type": "invalid",
	})
	if err == nil { t.Error("should error on invalid engagement_type") }
}

func TestEmailEngagement_ValidateConfig_ValidEngagementTypes(t *testing.T) {
	tests := []string{"opens", "clicks", "sends", "all"}
	for _, engType := range tests {
		err := (&EmailEngagement{}).ValidateConfig(map[string]interface{}{
			"engagement_type": engType,
		})
		if err != nil { t.Errorf("should accept %s: %v", engType, err) }
	}
}

func TestEmailEngagement_Execute_HighlyEngagedContact(t *testing.T) {
	recent := time.Now().Add(-5 * 24 * time.Hour).Format(time.RFC3339)
	mock := &mockConnectorForEmailEngagement{
		fieldValues: map[string]interface{}{
			"open_count":      10.0,
			"click_count":     3.0,
			"last_open_date":  recent,
			"last_click_date": recent,
		},
	}

	output, err := (&EmailEngagement{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"engagement_type": "all",
			"lookback_days":   30.0,
			"thresholds": map[string]interface{}{
				"highly_engaged": 5.0,
				"engaged":        2.0,
			},
			"tags": map[string]interface{}{
				"highly_engaged_tag": "tag_highly",
				"engaged_tag":        "tag_engaged",
				"disengaged_tag":     "tag_disengaged",
			},
		},
		Connector: mock,
	})
	if err != nil { t.Fatal(err) }
	if !output.Success { t.Error("should succeed") }

	// 10 opens + (3 clicks * 2) = 16, threshold 5 => highly_engaged
	if output.ModifiedData["engagement_level"] != "highly_engaged" {
		t.Errorf("expected highly_engaged, got %v", output.ModifiedData["engagement_level"])
	}
	if output.ModifiedData["engagement_score"] != 16 {
		t.Errorf("expected score 16, got %v", output.ModifiedData["engagement_score"])
	}

	// Should apply highly_engaged tag, remove others
	if len(mock.tagsApplied) != 1 || mock.tagsApplied[0] != "tag_highly" {
		t.Errorf("expected tag_highly applied, got %v", mock.tagsApplied)
	}
}

func TestEmailEngagement_Execute_EngagedContact(t *testing.T) {
	recent := time.Now().Add(-10 * 24 * time.Hour).Format(time.RFC3339)
	mock := &mockConnectorForEmailEngagement{
		fieldValues: map[string]interface{}{
			"open_count":     3.0,
			"click_count":    0.0,
			"last_open_date": recent,
		},
	}

	output, err := (&EmailEngagement{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"thresholds": map[string]interface{}{
				"highly_engaged": 5.0,
				"engaged":        2.0,
			},
			"tags": map[string]interface{}{
				"engaged_tag": "tag_engaged",
			},
		},
		Connector: mock,
	})
	if err != nil { t.Fatal(err) }

	// 3 opens, threshold 2 => engaged
	if output.ModifiedData["engagement_level"] != "engaged" {
		t.Errorf("expected engaged, got %v", output.ModifiedData["engagement_level"])
	}
}

func TestEmailEngagement_Execute_DisengagedContact(t *testing.T) {
	old := time.Now().Add(-60 * 24 * time.Hour).Format(time.RFC3339)
	mock := &mockConnectorForEmailEngagement{
		fieldValues: map[string]interface{}{
			"open_count":     10.0,
			"last_open_date": old, // outside lookback window
		},
	}

	output, err := (&EmailEngagement{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"lookback_days": 30.0,
			"tags": map[string]interface{}{
				"disengaged_tag": "tag_disengaged",
			},
		},
		Connector: mock,
	})
	if err != nil { t.Fatal(err) }

	// Score is 10 but last engagement is old => disengaged
	if output.ModifiedData["engagement_level"] != "disengaged" {
		t.Errorf("expected disengaged, got %v", output.ModifiedData["engagement_level"])
	}
}

func TestEmailEngagement_Execute_OpensOnly(t *testing.T) {
	recent := time.Now().Add(-5 * 24 * time.Hour).Format(time.RFC3339)
	mock := &mockConnectorForEmailEngagement{
		fieldValues: map[string]interface{}{
			"open_count":     5.0,
			"last_open_date": recent,
		},
	}

	output, err := (&EmailEngagement{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"engagement_type": "opens",
		},
		Connector: mock,
	})
	if err != nil { t.Fatal(err) }

	if output.ModifiedData["engagement_score"] != 5 {
		t.Errorf("expected score 5, got %v", output.ModifiedData["engagement_score"])
	}
}

func TestEmailEngagement_Execute_ClicksWeighted(t *testing.T) {
	recent := time.Now().Add(-5 * 24 * time.Hour).Format(time.RFC3339)
	mock := &mockConnectorForEmailEngagement{
		fieldValues: map[string]interface{}{
			"open_count":      2.0,
			"click_count":     3.0,
			"last_click_date": recent,
		},
	}

	output, err := (&EmailEngagement{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"engagement_type": "all",
		},
		Connector: mock,
	})
	if err != nil { t.Fatal(err) }

	// 2 opens + (3 clicks * 2) = 8
	if output.ModifiedData["engagement_score"] != 8 {
		t.Errorf("expected score 8 (clicks weighted x2), got %v", output.ModifiedData["engagement_score"])
	}
}

func TestEmailEngagement_Execute_ScoreFieldUpdate(t *testing.T) {
	recent := time.Now().Add(-5 * 24 * time.Hour).Format(time.RFC3339)
	mock := &mockConnectorForEmailEngagement{
		fieldValues: map[string]interface{}{
			"open_count":     3.0,
			"last_open_date": recent,
		},
	}

	output, err := (&EmailEngagement{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"score_field": "EngagementScore",
		},
		Connector: mock,
	})
	if err != nil { t.Fatal(err) }

	if mock.fieldValues["EngagementScore"] != 3 {
		t.Errorf("expected EngagementScore=3, got %v", mock.fieldValues["EngagementScore"])
	}

	// Check action
	found := false
	for _, action := range output.Actions {
		if action.Type == "field_updated" && action.Target == "EngagementScore" {
			found = true
			break
		}
	}
	if !found { t.Error("expected field_updated action for score_field") }
}

func TestEmailEngagement_Execute_LastEngagementFieldUpdate(t *testing.T) {
	recent := time.Now().Add(-5 * 24 * time.Hour)
	recentStr := recent.Format(time.RFC3339)
	mock := &mockConnectorForEmailEngagement{
		fieldValues: map[string]interface{}{
			"open_count":     2.0,
			"last_open_date": recentStr,
		},
	}

	output, err := (&EmailEngagement{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"last_engagement_field": "LastEngagement",
		},
		Connector: mock,
	})
	if err != nil { t.Fatal(err) }
	if !output.Success { t.Error("should succeed") }

	if mock.fieldValues["LastEngagement"] != recentStr {
		t.Errorf("expected LastEngagement=%s, got %v", recentStr, mock.fieldValues["LastEngagement"])
	}
}

func TestEmailEngagement_Execute_DefaultConfig(t *testing.T) {
	recent := time.Now().Add(-5 * 24 * time.Hour).Format(time.RFC3339)
	mock := &mockConnectorForEmailEngagement{
		fieldValues: map[string]interface{}{
			"open_count":     2.0,
			"last_open_date": recent,
		},
	}

	output, err := (&EmailEngagement{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config:    map[string]interface{}{}, // all defaults
		Connector: mock,
	})
	if err != nil { t.Fatal(err) }
	if !output.Success { t.Error("should succeed with default config") }
}

func TestEmailEngagement_Execute_TagRemovalForOtherLevels(t *testing.T) {
	recent := time.Now().Add(-5 * 24 * time.Hour).Format(time.RFC3339)
	mock := &mockConnectorForEmailEngagement{
		fieldValues: map[string]interface{}{
			"open_count":     10.0,
			"click_count":    3.0,
			"last_open_date": recent,
		},
	}

	output, err := (&EmailEngagement{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"thresholds": map[string]interface{}{
				"highly_engaged": 5.0,
				"engaged":        2.0,
			},
			"tags": map[string]interface{}{
				"highly_engaged_tag": "tag_highly",
				"engaged_tag":        "tag_engaged",
				"disengaged_tag":     "tag_disengaged",
			},
		},
		Connector: mock,
	})
	if err != nil { t.Fatal(err) }
	if !output.Success { t.Error("should succeed") }

	// Should apply highly_engaged, remove engaged and disengaged
	if len(mock.tagsApplied) != 1 || mock.tagsApplied[0] != "tag_highly" {
		t.Errorf("expected tag_highly applied, got %v", mock.tagsApplied)
	}
	if len(mock.tagsRemoved) != 2 {
		t.Errorf("expected 2 tags removed, got %d", len(mock.tagsRemoved))
	}
}

func TestEmailEngagement_Execute_NoEngagementData(t *testing.T) {
	mock := &mockConnectorForEmailEngagement{
		fieldValues: map[string]interface{}{},
	}

	output, err := (&EmailEngagement{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config:    map[string]interface{}{},
		Connector: mock,
	})
	if err != nil { t.Fatal(err) }

	// No data => disengaged, score 0
	if output.ModifiedData["engagement_level"] != "disengaged" {
		t.Errorf("expected disengaged with no data, got %v", output.ModifiedData["engagement_level"])
	}
	if output.ModifiedData["engagement_score"] != 0 {
		t.Errorf("expected score 0, got %v", output.ModifiedData["engagement_score"])
	}
}

func TestEmailEngagement_Execute_GetFieldError(t *testing.T) {
	mock := &mockConnectorForEmailEngagement{
		getFieldError: map[string]error{
			"open_count": fmt.Errorf("API error"),
		},
	}

	output, err := (&EmailEngagement{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config:    map[string]interface{}{},
		Connector: mock,
	})
	if err != nil { t.Fatal(err) }

	// Should still succeed with partial data
	if !output.Success { t.Error("should succeed even with field errors") }
}

func TestEmailEngagement_Execute_SetFieldError(t *testing.T) {
	recent := time.Now().Add(-5 * 24 * time.Hour).Format(time.RFC3339)
	mock := &mockConnectorForEmailEngagement{
		fieldValues: map[string]interface{}{
			"open_count":     3.0,
			"last_open_date": recent,
		},
		setFieldError: map[string]error{
			"EngagementScore": fmt.Errorf("field update error"),
		},
	}

	output, err := (&EmailEngagement{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"score_field": "EngagementScore",
		},
		Connector: mock,
	})
	if err != nil { t.Fatal(err) }

	// Should still succeed, just log error
	if !output.Success { t.Error("should succeed even with set field error") }

	// Check logs mention the error
	foundLog := false
	for _, log := range output.Logs {
		if len(log) > 0 && log[:6] == "Failed" {
			foundLog = true
			break
		}
	}
	if !foundLog { t.Error("expected error log for failed field set") }
}

func TestEmailEngagement_Execute_ApplyTagError(t *testing.T) {
	recent := time.Now().Add(-5 * 24 * time.Hour).Format(time.RFC3339)
	mock := &mockConnectorForEmailEngagement{
		fieldValues: map[string]interface{}{
			"open_count":     3.0,
			"last_open_date": recent,
		},
		applyTagError: map[string]error{
			"tag_engaged": fmt.Errorf("tag apply error"),
		},
	}

	output, err := (&EmailEngagement{}).Execute(context.Background(), helpers.HelperInput{
		ContactID: "123",
		Config: map[string]interface{}{
			"tags": map[string]interface{}{
				"engaged_tag": "tag_engaged",
			},
		},
		Connector: mock,
	})
	if err != nil { t.Fatal(err) }

	// Should still succeed, just log error
	if !output.Success { t.Error("should succeed even with tag apply error") }
}
