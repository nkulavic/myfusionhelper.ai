package automation

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/myfusionhelper/api/internal/connectors"
	"github.com/myfusionhelper/api/internal/helpers"
)

// Mock connector for testing contact_field mode
type mockConnectorForTimer struct {
	fieldValues map[string]interface{}
}

func (m *mockConnectorForTimer) GetContactFieldValue(ctx context.Context, contactID, fieldKey string) (interface{}, error) {
	if val, ok := m.fieldValues[fieldKey]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("field not found: %s", fieldKey)
}

func (m *mockConnectorForTimer) GetContacts(ctx context.Context, opts connectors.QueryOptions) (*connectors.ContactList, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForTimer) GetContact(ctx context.Context, contactID string) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForTimer) CreateContact(ctx context.Context, input connectors.CreateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForTimer) UpdateContact(ctx context.Context, contactID string, updates connectors.UpdateContactInput) (*connectors.NormalizedContact, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForTimer) DeleteContact(ctx context.Context, contactID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForTimer) GetTags(ctx context.Context) ([]connectors.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForTimer) ApplyTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForTimer) RemoveTag(ctx context.Context, contactID, tagID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForTimer) GetCustomFields(ctx context.Context) ([]connectors.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockConnectorForTimer) SetContactFieldValue(ctx context.Context, contactID, fieldKey string, value interface{}) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForTimer) TriggerAutomation(ctx context.Context, contactID, automationID string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForTimer) AchieveGoal(ctx context.Context, contactID, goalName, integration string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForTimer) TestConnection(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}
func (m *mockConnectorForTimer) GetMetadata() connectors.ConnectorMetadata {
	return connectors.ConnectorMetadata{}
}
func (m *mockConnectorForTimer) GetCapabilities() []connectors.Capability {
	return nil
}
func (m *mockConnectorForTimer) SetOptInStatus(ctx context.Context, contactID string, optIn bool, reason string) error {
	return nil
}

// Test helper metadata
func TestCountdownTimer_Metadata(t *testing.T) {
	helper := &CountdownTimer{}

	if helper.GetName() != "Countdown Timer" {
		t.Errorf("Expected name 'Countdown Timer', got '%s'", helper.GetName())
	}
	if helper.GetType() != "countdown_timer" {
		t.Errorf("Expected type 'countdown_timer', got '%s'", helper.GetType())
	}
	if helper.GetCategory() != "automation" {
		t.Errorf("Expected category 'automation', got '%s'", helper.GetCategory())
	}
	if helper.RequiresCRM() {
		t.Error("Expected RequiresCRM to be false")
	}

	schema := helper.GetConfigSchema()
	if schema == nil {
		t.Fatal("Expected config schema, got nil")
	}
}

// Test validation - missing timerType
func TestCountdownTimer_ValidateConfig_MissingTimerType(t *testing.T) {
	helper := &CountdownTimer{}
	config := map[string]interface{}{}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected validation error for missing timerType")
	}
	if !strings.Contains(err.Error(), "timerType") {
		t.Errorf("Expected error about timerType, got: %v", err)
	}
}

// Test validation - standard mode missing endTime
func TestCountdownTimer_ValidateConfig_StandardMissingEndTime(t *testing.T) {
	helper := &CountdownTimer{}
	config := map[string]interface{}{
		"timerType": "standard",
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected validation error for missing endTime")
	}
	if !strings.Contains(err.Error(), "endTime") {
		t.Errorf("Expected error about endTime, got: %v", err)
	}
}

// Test validation - standard mode invalid endTime format
func TestCountdownTimer_ValidateConfig_StandardInvalidEndTime(t *testing.T) {
	helper := &CountdownTimer{}
	config := map[string]interface{}{
		"timerType": "standard",
		"endTime":   "not-a-datetime",
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected validation error for invalid endTime format")
	}
}

// Test validation - standard mode valid
func TestCountdownTimer_ValidateConfig_StandardValid(t *testing.T) {
	helper := &CountdownTimer{}
	config := map[string]interface{}{
		"timerType": "standard",
		"endTime":   "2025-12-31T23:59:59Z",
	}

	err := helper.ValidateConfig(config)
	if err != nil {
		t.Errorf("Expected no validation error, got: %v", err)
	}
}

// Test validation - contact_field mode missing contactField
func TestCountdownTimer_ValidateConfig_ContactFieldMissing(t *testing.T) {
	helper := &CountdownTimer{}
	config := map[string]interface{}{
		"timerType": "contact_field",
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected validation error for missing contactField")
	}
	if !strings.Contains(err.Error(), "contactField") {
		t.Errorf("Expected error about contactField, got: %v", err)
	}
}

// Test validation - contact_field mode valid
func TestCountdownTimer_ValidateConfig_ContactFieldValid(t *testing.T) {
	helper := &CountdownTimer{}
	config := map[string]interface{}{
		"timerType":    "contact_field",
		"contactField": "deadline_date",
	}

	err := helper.ValidateConfig(config)
	if err != nil {
		t.Errorf("Expected no validation error, got: %v", err)
	}
}

// Test validation - evergreen mode no duration
func TestCountdownTimer_ValidateConfig_EvergreenNoDuration(t *testing.T) {
	helper := &CountdownTimer{}
	config := map[string]interface{}{
		"timerType": "evergreen",
		"addDays":   float64(0),
		"addHours":  float64(0),
		"addMinutes": float64(0),
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected validation error for evergreen with zero duration")
	}
	if !strings.Contains(err.Error(), "greater than 0") {
		t.Errorf("Expected error about duration, got: %v", err)
	}
}

// Test validation - evergreen mode valid
func TestCountdownTimer_ValidateConfig_EvergreenValid(t *testing.T) {
	helper := &CountdownTimer{}
	config := map[string]interface{}{
		"timerType":  "evergreen",
		"addDays":    float64(7),
		"addHours":   float64(0),
		"addMinutes": float64(0),
	}

	err := helper.ValidateConfig(config)
	if err != nil {
		t.Errorf("Expected no validation error, got: %v", err)
	}
}

// Test validation - invalid timerType
func TestCountdownTimer_ValidateConfig_InvalidTimerType(t *testing.T) {
	helper := &CountdownTimer{}
	config := map[string]interface{}{
		"timerType": "unknown",
	}

	err := helper.ValidateConfig(config)
	if err == nil {
		t.Error("Expected validation error for invalid timerType")
	}
	if !strings.Contains(err.Error(), "invalid timerType") {
		t.Errorf("Expected error about invalid timerType, got: %v", err)
	}
}

// Test execution - standard mode
func TestCountdownTimer_Execute_StandardMode(t *testing.T) {
	// Set JWT secret for testing (raw secret, not base64)
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	helper := &CountdownTimer{}
	deadline := time.Now().Add(24 * time.Hour)

	config := map[string]interface{}{
		"timerType":       "standard",
		"endTime":         deadline.Format(time.RFC3339),
		"backgroundColor": "#FF0000",
		"digitColor":      "#00FF00",
		"labelColor":      "#0000FF",
		"transparentBg":   false,
	}

	input := helpers.HelperInput{
		ContactID: "contact-123",
		AccountID: "account-456",
		HelperID:  "helper-789",
		Config:    config,
		UserID:    "user-999",
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	timerURL, ok := output.ModifiedData["timer_url"].(string)
	if !ok || timerURL == "" {
		t.Error("Expected timer_url in ModifiedData")
	}

	// Verify URL contains JWT token
	if !strings.Contains(timerURL, "/timer?t=") {
		t.Errorf("Expected timer URL format, got: %s", timerURL)
	}

	// Extract and validate JWT
	tokenPart := strings.Split(timerURL, "?t=")
	if len(tokenPart) != 2 {
		t.Fatalf("Expected token in URL, got: %s", timerURL)
	}

	token, err := jwt.Parse(tokenPart[1], func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret-key"), nil
	})
	if err != nil {
		t.Fatalf("Failed to parse JWT: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("Failed to get JWT claims")
	}

	if claims["contact_id"] != "contact-123" {
		t.Errorf("Expected contact_id=contact-123, got: %v", claims["contact_id"])
	}
	if claims["bg_color"] != "#FF0000" {
		t.Errorf("Expected bg_color=#FF0000, got: %v", claims["bg_color"])
	}
	if claims["digit_color"] != "#00FF00" {
		t.Errorf("Expected digit_color=#00FF00, got: %v", claims["digit_color"])
	}
	if claims["label_color"] != "#0000FF" {
		t.Errorf("Expected label_color=#0000FF, got: %v", claims["label_color"])
	}
	if claims["transparent_bg"] != false {
		t.Errorf("Expected transparent_bg=false, got: %v", claims["transparent_bg"])
	}
}

// Test execution - contact_field mode
func TestCountdownTimer_Execute_ContactFieldMode(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	helper := &CountdownTimer{}
	deadline := time.Now().Add(48 * time.Hour)

	mockConn := &mockConnectorForTimer{
		fieldValues: map[string]interface{}{
			"webinar_date": deadline.Format(time.RFC3339),
		},
	}

	config := map[string]interface{}{
		"timerType":    "contact_field",
		"contactField": "webinar_date",
	}

	input := helpers.HelperInput{
		ContactID: "contact-abc",
		AccountID: "account-def",
		HelperID:  "helper-ghi",
		Config:    config,
		Connector: mockConn,
		UserID:    "user-jkl",
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	timerURL, ok := output.ModifiedData["timer_url"].(string)
	if !ok || timerURL == "" {
		t.Error("Expected timer_url in ModifiedData")
	}

	expiresAt, ok := output.ModifiedData["expires_at"].(string)
	if !ok || expiresAt == "" {
		t.Error("Expected expires_at in ModifiedData")
	}

	// Verify expires_at matches the field value
	parsedExpiry, err := time.Parse(time.RFC3339, expiresAt)
	if err != nil {
		t.Fatalf("Failed to parse expires_at: %v", err)
	}

	if parsedExpiry.Unix() != deadline.Unix() {
		t.Errorf("Expected expires_at=%d, got=%d", deadline.Unix(), parsedExpiry.Unix())
	}
}

// Test execution - contact_field mode with missing field
func TestCountdownTimer_Execute_ContactFieldModeMissingField(t *testing.T) {
	helper := &CountdownTimer{}

	mockConn := &mockConnectorForTimer{
		fieldValues: map[string]interface{}{},
	}

	config := map[string]interface{}{
		"timerType":    "contact_field",
		"contactField": "missing_field",
	}

	input := helpers.HelperInput{
		ContactID: "contact-abc",
		Config:    config,
		Connector: mockConn,
	}

	output, err := helper.Execute(context.Background(), input)
	if err == nil {
		t.Error("Expected error for missing field")
	}
	if output.Success {
		t.Error("Expected success=false")
	}
}

// Test execution - contact_field mode with no connector
func TestCountdownTimer_Execute_ContactFieldModeNoConnector(t *testing.T) {
	helper := &CountdownTimer{}

	config := map[string]interface{}{
		"timerType":    "contact_field",
		"contactField": "deadline",
	}

	input := helpers.HelperInput{
		ContactID: "contact-123",
		Config:    config,
		Connector: nil,
	}

	output, err := helper.Execute(context.Background(), input)
	if err == nil {
		t.Error("Expected error for missing connector")
	}
	if !strings.Contains(output.Message, "connector required") {
		t.Errorf("Expected error message about connector, got: %s", output.Message)
	}
}

// Test execution - evergreen mode (without DynamoDB)
func TestCountdownTimer_Execute_EvergreenMode(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	// Unset DynamoDB table to test local execution
	os.Unsetenv("COUNTDOWN_TIMERS_TABLE")

	helper := &CountdownTimer{}

	config := map[string]interface{}{
		"timerType":  "evergreen",
		"addDays":    float64(3),
		"addHours":   float64(12),
		"addMinutes": float64(30),
	}

	input := helpers.HelperInput{
		ContactID: "contact-xyz",
		AccountID: "account-uvw",
		HelperID:  "helper-rst",
		Config:    config,
		UserID:    "user-opq",
	}

	beforeExec := time.Now()
	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !output.Success {
		t.Error("Expected success=true")
	}

	timerURL, ok := output.ModifiedData["timer_url"].(string)
	if !ok || timerURL == "" {
		t.Error("Expected timer_url in ModifiedData")
	}

	expiresAt, ok := output.ModifiedData["expires_at"].(string)
	if !ok || expiresAt == "" {
		t.Error("Expected expires_at in ModifiedData")
	}

	// Verify expires_at is approximately 3.5 days from now
	parsedExpiry, err := time.Parse(time.RFC3339, expiresAt)
	if err != nil {
		t.Fatalf("Failed to parse expires_at: %v", err)
	}

	expectedDuration := 3*24*time.Hour + 12*time.Hour + 30*time.Minute
	expectedExpiry := beforeExec.Add(expectedDuration)

	// Allow 5 second tolerance for test execution time
	diff := parsedExpiry.Sub(expectedExpiry)
	if diff < -5*time.Second || diff > 5*time.Second {
		t.Errorf("Expected expires_at around %s, got %s (diff: %v)", expectedExpiry.Format(time.RFC3339), parsedExpiry.Format(time.RFC3339), diff)
	}
}

// Test execution - transparent background
func TestCountdownTimer_Execute_TransparentBackground(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	helper := &CountdownTimer{}
	deadline := time.Now().Add(24 * time.Hour)

	config := map[string]interface{}{
		"timerType":     "standard",
		"endTime":       deadline.Format(time.RFC3339),
		"transparentBg": true,
	}

	input := helpers.HelperInput{
		ContactID: "contact-123",
		AccountID: "account-456",
		HelperID:  "helper-789",
		Config:    config,
		UserID:    "user-999",
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	timerURL := output.ModifiedData["timer_url"].(string)
	tokenPart := strings.Split(timerURL, "?t=")
	token, _ := jwt.Parse(tokenPart[1], func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret-key"), nil
	})

	claims := token.Claims.(jwt.MapClaims)
	if claims["transparent_bg"] != true {
		t.Errorf("Expected transparent_bg=true, got: %v", claims["transparent_bg"])
	}
}

// Test execution - default colors
func TestCountdownTimer_Execute_DefaultColors(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	helper := &CountdownTimer{}
	deadline := time.Now().Add(24 * time.Hour)

	config := map[string]interface{}{
		"timerType": "standard",
		"endTime":   deadline.Format(time.RFC3339),
		// No colors specified - should use defaults
	}

	input := helpers.HelperInput{
		ContactID: "contact-123",
		AccountID: "account-456",
		HelperID:  "helper-789",
		Config:    config,
		UserID:    "user-999",
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	timerURL := output.ModifiedData["timer_url"].(string)
	tokenPart := strings.Split(timerURL, "?t=")
	token, _ := jwt.Parse(tokenPart[1], func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret-key"), nil
	})

	claims := token.Claims.(jwt.MapClaims)
	if claims["bg_color"] != "#000000" {
		t.Errorf("Expected default bg_color=#000000, got: %v", claims["bg_color"])
	}
	if claims["digit_color"] != "#FFFFFF" {
		t.Errorf("Expected default digit_color=#FFFFFF, got: %v", claims["digit_color"])
	}
	if claims["label_color"] != "#CCCCCC" {
		t.Errorf("Expected default label_color=#CCCCCC, got: %v", claims["label_color"])
	}
}

// Test action logging
func TestCountdownTimer_Execute_ActionLogging(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	helper := &CountdownTimer{}
	deadline := time.Now().Add(24 * time.Hour)

	config := map[string]interface{}{
		"timerType": "standard",
		"endTime":   deadline.Format(time.RFC3339),
	}

	input := helpers.HelperInput{
		ContactID: "contact-123",
		AccountID: "account-456",
		HelperID:  "helper-789",
		Config:    config,
		UserID:    "user-999",
	}

	output, err := helper.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(output.Actions) == 0 {
		t.Error("Expected at least one action logged")
	}

	foundTimerGenerated := false
	for _, action := range output.Actions {
		if action.Type == "timer_generated" && action.Target == "contact-123" {
			foundTimerGenerated = true
		}
	}

	if !foundTimerGenerated {
		t.Error("Expected timer_generated action to be logged")
	}
}
