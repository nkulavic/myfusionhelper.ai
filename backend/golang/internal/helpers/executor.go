package helpers

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/myfusionhelper/api/internal/connectors"
)

// ExecutionRequest represents a request to execute a helper
type ExecutionRequest struct {
	HelperType   string                                  `json:"helper_type"`
	ContactID    string                                  `json:"contact_id"`
	Config       map[string]interface{}                   `json:"config"`
	Input        map[string]interface{}                   `json:"input"`        // Per-execution data from POST body
	QueryParams  map[string]string                        `json:"query_params"` // Query string parameters from request
	UserID       string                                  `json:"user_id"`
	AccountID    string                                  `json:"account_id"`
	HelperID     string                                  `json:"helper_id"`
	ConnectionID string                                  `json:"connection_id"`
	ServiceAuths map[string]*connectors.ConnectorConfig   `json:"-"` // pre-loaded service connection credentials
	APIKey       string                                  `json:"-"` // Original x-api-key header for relay helpers
}

// ExecutionResult represents the full result of a helper execution
type ExecutionResult struct {
	Success      bool                   `json:"success"`
	Output       *HelperOutput          `json:"output,omitempty"`
	Error        string                 `json:"error,omitempty"`
	HelperType   string                 `json:"helper_type"`
	ContactID    string                 `json:"contact_id"`
	DurationMs   int64                  `json:"duration_ms"`
	ExecutedAt   time.Time              `json:"executed_at"`
}

// Executor handles the execution of helpers
type Executor struct{}

// NewExecutor creates a new helper executor
func NewExecutor() *Executor {
	return &Executor{}
}

// Execute runs a helper with the given request and connector
func (e *Executor) Execute(ctx context.Context, req ExecutionRequest, connector connectors.CRMConnector) (*ExecutionResult, error) {
	start := time.Now()

	result := &ExecutionResult{
		HelperType: req.HelperType,
		ContactID:  req.ContactID,
		ExecutedAt: start,
	}

	// Look up the helper implementation
	helper, err := NewHelper(req.HelperType)
	if err != nil {
		result.Error = fmt.Sprintf("unknown helper type: %s", req.HelperType)
		result.DurationMs = time.Since(start).Milliseconds()
		return result, err
	}

	// Validate config
	if err := helper.ValidateConfig(req.Config); err != nil {
		result.Error = fmt.Sprintf("invalid config: %v", err)
		result.DurationMs = time.Since(start).Milliseconds()
		return result, err
	}

	// Check if CRM connector is required but not provided
	if helper.RequiresCRM() && connector == nil {
		result.Error = "helper requires a CRM connection but none was provided"
		result.DurationMs = time.Since(start).Milliseconds()
		return result, fmt.Errorf("%s", result.Error)
	}

	// Fetch contact data if connector is available and contact ID is provided
	var contactData *connectors.NormalizedContact
	if connector != nil && req.ContactID != "" {
		contact, err := connector.GetContact(ctx, req.ContactID)
		if err != nil {
			log.Printf("Warning: Failed to fetch contact %s: %v", req.ContactID, err)
			// Don't fail - some helpers might not need full contact data
		} else {
			contactData = contact
		}
	}

	// Build helper input
	input := HelperInput{
		ContactID:    req.ContactID,
		ContactData:  contactData,
		Config:       req.Config,
		Input:        req.Input,
		QueryParams:  req.QueryParams,
		Connector:    connector,
		ServiceAuths: req.ServiceAuths,
		UserID:       req.UserID,
		AccountID:    req.AccountID,
		HelperID:     req.HelperID,
		APIKey:       req.APIKey,
	}

	// Execute the helper
	output, err := helper.Execute(ctx, input)
	result.DurationMs = time.Since(start).Milliseconds()

	if err != nil {
		result.Error = err.Error()
		result.Output = output
		return result, err
	}

	result.Success = output.Success
	result.Output = output
	return result, nil
}
