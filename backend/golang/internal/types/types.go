package types

import (
	"time"
)

// ========== USER & ACCOUNT TYPES ==========

// User represents a user in the system
type User struct {
	UserID                  string                   `json:"user_id" dynamodbav:"user_id"`
	CognitoUserID           string                   `json:"cognito_user_id" dynamodbav:"cognito_user_id"`
	Email                   string                   `json:"email" dynamodbav:"email"`
	Name                    string                   `json:"name" dynamodbav:"name"`
	PhoneNumber             string                   `json:"phone_number,omitempty" dynamodbav:"phone_number,omitempty"`
	Company                 string                   `json:"company,omitempty" dynamodbav:"company,omitempty"`
	Status                  string                   `json:"status" dynamodbav:"status"`
	CurrentAccountID        string                   `json:"current_account_id" dynamodbav:"current_account_id"`
	OnboardingComplete      bool                     `json:"onboarding_complete" dynamodbav:"onboarding_complete"`
	NotificationPreferences *NotificationPreferences `json:"notification_preferences,omitempty" dynamodbav:"notification_preferences,omitempty"`
	CreatedAt               time.Time                `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt               time.Time                `json:"updated_at" dynamodbav:"updated_at"`
	LastLoginAt             *time.Time               `json:"last_login_at,omitempty" dynamodbav:"last_login_at,omitempty"`
}

// NotificationPreferences represents user notification settings
type NotificationPreferences struct {
	ExecutionFailures bool   `json:"execution_failures" dynamodbav:"execution_failures"`
	ConnectionIssues  bool   `json:"connection_issues" dynamodbav:"connection_issues"`
	UsageAlerts       bool   `json:"usage_alerts" dynamodbav:"usage_alerts"`
	WeeklySummary     bool   `json:"weekly_summary" dynamodbav:"weekly_summary"`
	NewFeatures       bool   `json:"new_features" dynamodbav:"new_features"`
	TeamActivity      bool   `json:"team_activity" dynamodbav:"team_activity"`
	RealtimeStatus    bool   `json:"realtime_status" dynamodbav:"realtime_status"`
	AiInsights        bool   `json:"ai_insights" dynamodbav:"ai_insights"`
	SystemMaintenance bool   `json:"system_maintenance" dynamodbav:"system_maintenance"`
	WebhookURL        string `json:"webhook_url,omitempty" dynamodbav:"webhook_url,omitempty"`
}

// Account represents a billing entity / workspace
type Account struct {
	AccountID        string          `json:"account_id" dynamodbav:"account_id"`
	OwnerUserID      string          `json:"owner_user_id" dynamodbav:"owner_user_id"`
	CreatedByUserID  string          `json:"created_by_user_id" dynamodbav:"created_by_user_id"`
	Name             string          `json:"name" dynamodbav:"name"`
	Company          string          `json:"company" dynamodbav:"company"`
	Plan             string          `json:"plan" dynamodbav:"plan"`
	Status           string          `json:"status" dynamodbav:"status"`
	StripeCustomerID string          `json:"stripe_customer_id,omitempty" dynamodbav:"stripe_customer_id,omitempty"`
	Settings         AccountSettings `json:"settings" dynamodbav:"settings"`
	Usage            AccountUsage    `json:"usage" dynamodbav:"usage"`
	CreatedAt        time.Time       `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at" dynamodbav:"updated_at"`
}

// AccountSettings represents configurable account settings
type AccountSettings struct {
	MaxHelpers          int    `json:"max_helpers" dynamodbav:"max_helpers"`
	MaxConnections      int    `json:"max_connections" dynamodbav:"max_connections"`
	MaxAPIKeys          int    `json:"max_api_keys" dynamodbav:"max_api_keys"`
	MaxTeamMembers      int    `json:"max_team_members" dynamodbav:"max_team_members"`
	MaxExecutions       int    `json:"max_executions" dynamodbav:"max_executions"`
	WebhooksEnabled     bool   `json:"webhooks_enabled" dynamodbav:"webhooks_enabled"`
	StripeMeteredItemID string `json:"stripe_metered_item_id,omitempty" dynamodbav:"stripe_metered_item_id,omitempty"`
}

// AccountUsage represents current usage metrics for an account
type AccountUsage struct {
	Helpers             int `json:"helpers" dynamodbav:"helpers"`
	Connections         int `json:"connections" dynamodbav:"connections"`
	APIKeys             int `json:"api_keys" dynamodbav:"api_keys"`
	TeamMembers         int `json:"team_members" dynamodbav:"team_members"`
	MonthlyExecutions   int `json:"monthly_executions" dynamodbav:"monthly_executions"`
	MonthlyAPIRequests  int `json:"monthly_api_requests" dynamodbav:"monthly_api_requests"`
}

// UserAccount represents the many-to-many relationship between users and accounts
type UserAccount struct {
	UserID      string      `json:"user_id" dynamodbav:"user_id"`
	AccountID   string      `json:"account_id" dynamodbav:"account_id"`
	Role        string      `json:"role" dynamodbav:"role"`
	Status      string      `json:"status" dynamodbav:"status"`
	Permissions Permissions `json:"permissions" dynamodbav:"permissions"`
	LinkedAt    string      `json:"linked_at" dynamodbav:"linked_at"`
	UpdatedAt   string      `json:"updated_at" dynamodbav:"updated_at"`
}

// ========== AUTH CONTEXT TYPES ==========

// AuthContext holds the authenticated user's context
type AuthContext struct {
	UserID            string          `json:"user_id"`
	AccountID         string          `json:"account_id"`
	Email             string          `json:"email"`
	Role              string          `json:"role"`
	Permissions       Permissions     `json:"permissions"`
	AvailableAccounts []AccountAccess `json:"available_accounts"`
}

// Permissions defines granular user permissions within an account
type Permissions struct {
	CanManageHelpers     bool `json:"can_manage_helpers" dynamodbav:"can_manage_helpers"`
	CanExecuteHelpers    bool `json:"can_execute_helpers" dynamodbav:"can_execute_helpers"`
	CanManageConnections bool `json:"can_manage_connections" dynamodbav:"can_manage_connections"`
	CanManageTeam        bool `json:"can_manage_team" dynamodbav:"can_manage_team"`
	CanManageBilling     bool `json:"can_manage_billing" dynamodbav:"can_manage_billing"`
	CanViewAnalytics     bool `json:"can_view_analytics" dynamodbav:"can_view_analytics"`
	CanManageAPIKeys     bool `json:"can_manage_api_keys" dynamodbav:"can_manage_api_keys"`
}

// AccountAccess represents a user's access to an account
type AccountAccess struct {
	AccountID   string      `json:"account_id"`
	AccountName string      `json:"account_name"`
	Role        string      `json:"role"`
	Permissions Permissions `json:"permissions"`
	IsCurrent   bool        `json:"is_current"`
}

// ========== PLATFORM TYPES ==========

// Platform represents a CRM service provider (Keap, GoHighLevel, ActiveCampaign, etc.)
type Platform struct {
	PlatformID       string              `json:"platform_id" dynamodbav:"platform_id"`
	Name             string              `json:"name" dynamodbav:"name"`
	Slug             string              `json:"slug" dynamodbav:"slug"`
	Category         string              `json:"category" dynamodbav:"category"`
	Types            []string            `json:"types" dynamodbav:"types"`
	Description      string              `json:"description" dynamodbav:"description"`
	Status           string              `json:"status" dynamodbav:"status"`
	Version          string              `json:"version" dynamodbav:"version"`
	LogoURL          string              `json:"logo_url" dynamodbav:"logo_url"`
	DocumentationURL string              `json:"documentation_url" dynamodbav:"documentation_url"`
	OAuth            *OAuthConfiguration `json:"oauth,omitempty" dynamodbav:"oauth,omitempty"`
	APIConfig        APIConfiguration    `json:"api_config" dynamodbav:"api_config"`
	TestEndpoints    *TestEndpoints      `json:"test_endpoints,omitempty" dynamodbav:"test_endpoints,omitempty"`
	DisplayConfig    *DisplayConfig      `json:"display_config,omitempty" dynamodbav:"display_config,omitempty"`
	CredentialFields []CredentialField   `json:"credential_fields,omitempty" dynamodbav:"credential_fields,omitempty"`
	Capabilities     []string            `json:"capabilities" dynamodbav:"capabilities"`
	CreatedAt        time.Time           `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at" dynamodbav:"updated_at"`
}

// OAuthConfiguration represents OAuth settings for a platform
type OAuthConfiguration struct {
	AuthURL      string   `json:"auth_url" dynamodbav:"auth_url"`
	TokenURL     string   `json:"token_url" dynamodbav:"token_url"`
	UserInfoURL  string   `json:"user_info_url" dynamodbav:"user_info_url"`
	Scopes       []string `json:"scopes" dynamodbav:"scopes"`
	ResponseType string   `json:"response_type" dynamodbav:"response_type"`
}

// DisplayConfig holds UI metadata for rendering a platform in the frontend
type DisplayConfig struct {
	Color     string `json:"color" dynamodbav:"color"`
	Accent    string `json:"accent" dynamodbav:"accent"`
	Initial   string `json:"initial" dynamodbav:"initial"`
	ShortName string `json:"short_name,omitempty" dynamodbav:"short_name,omitempty"`
}

// CredentialField describes a single input field for platform credential entry
type CredentialField struct {
	Key         string `json:"key" dynamodbav:"key"`
	Label       string `json:"label" dynamodbav:"label"`
	Placeholder string `json:"placeholder" dynamodbav:"placeholder"`
	Hint        string `json:"hint,omitempty" dynamodbav:"hint,omitempty"`
	InputType   string `json:"input_type" dynamodbav:"input_type"`
	Required    bool   `json:"required" dynamodbav:"required"`
}

// APIConfiguration represents API configuration for a platform
type APIConfiguration struct {
	BaseURL         string            `json:"base_url" dynamodbav:"base_url"`
	AuthType        string            `json:"auth_type" dynamodbav:"auth_type"`
	TestEndpoint    string            `json:"test_endpoint" dynamodbav:"test_endpoint"`
	RateLimits      RateLimitConfig   `json:"rate_limits" dynamodbav:"rate_limits"`
	RequiredHeaders map[string]string `json:"required_headers" dynamodbav:"required_headers"`
	Version         string            `json:"version" dynamodbav:"version"`
}

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	RequestsPerSecond int `json:"requests_per_second" dynamodbav:"requests_per_second"`
	RequestsPerMinute int `json:"requests_per_minute" dynamodbav:"requests_per_minute"`
	RequestsPerHour   int `json:"requests_per_hour" dynamodbav:"requests_per_hour"`
	BurstLimit        int `json:"burst_limit" dynamodbav:"burst_limit"`
}

// TestEndpointConfig represents a test/validation endpoint for a platform
type TestEndpointConfig struct {
	URL        string `json:"url" dynamodbav:"url"`
	Method     string `json:"method" dynamodbav:"method"`
	AuthHeader string `json:"auth_header" dynamodbav:"auth_header"`
	AuthFormat string `json:"auth_format" dynamodbav:"auth_format"`
}

// TestEndpoints represents test endpoints for different auth types
type TestEndpoints struct {
	OAuth2 *TestEndpointConfig `json:"oauth2,omitempty" dynamodbav:"oauth2,omitempty"`
	Bearer *TestEndpointConfig `json:"bearer,omitempty" dynamodbav:"bearer,omitempty"`
	APIKey *TestEndpointConfig `json:"api_key,omitempty" dynamodbav:"api_key,omitempty"`
}

// ========== PLATFORM CONNECTION TYPES ==========

// PlatformConnection represents a user's authenticated connection to a CRM platform
type PlatformConnection struct {
	ConnectionID        string                 `json:"connection_id" dynamodbav:"connection_id"`
	AccountID           string                 `json:"account_id" dynamodbav:"account_id"`
	UserID              string                 `json:"user_id" dynamodbav:"user_id"`
	PlatformID          string                 `json:"platform_id" dynamodbav:"platform_id"`
	ExternalUserID      string                 `json:"external_user_id,omitempty" dynamodbav:"external_user_id,omitempty"`
	ExternalUserEmail   string                 `json:"external_user_email,omitempty" dynamodbav:"external_user_email,omitempty"`
	ExternalAppID       string                 `json:"external_app_id,omitempty" dynamodbav:"external_app_id,omitempty"`
	ExternalAppName     string                 `json:"external_app_name,omitempty" dynamodbav:"external_app_name,omitempty"`
	Name                string                 `json:"name" dynamodbav:"name"`
	Status              string                 `json:"status" dynamodbav:"status"`
	AuthType            string                 `json:"auth_type" dynamodbav:"auth_type"`
	AuthID              *string                `json:"auth_id,omitempty" dynamodbav:"auth_id,omitempty"`
	CredentialsMetadata map[string]interface{} `json:"credentials_metadata" dynamodbav:"credentials_metadata"`
	LastConnected       *time.Time             `json:"last_connected,omitempty" dynamodbav:"last_connected,omitempty"`
	CreatedAt           time.Time              `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at" dynamodbav:"updated_at"`
	ExpiresAt           *time.Time             `json:"expires_at,omitempty" dynamodbav:"expires_at,omitempty"`
	LastSyncedAt        *time.Time             `json:"last_synced_at,omitempty" dynamodbav:"last_synced_at,omitempty"`
	SyncStatus          string                 `json:"sync_status,omitempty" dynamodbav:"sync_status,omitempty"`
	SyncRecordCounts    map[string]int         `json:"sync_record_counts,omitempty" dynamodbav:"sync_record_counts,omitempty"`
}

// PlatformConnectionAuth represents authentication credentials for a platform connection
type PlatformConnectionAuth struct {
	AuthID        string                 `json:"auth_id" dynamodbav:"auth_id"`
	ConnectionID  string                 `json:"connection_id" dynamodbav:"connection_id"`
	AccountID     string                 `json:"account_id" dynamodbav:"account_id"`
	UserID        string                 `json:"user_id" dynamodbav:"user_id"`
	PlatformID    string                 `json:"platform_id" dynamodbav:"platform_id"`
	AuthType      string                 `json:"auth_type" dynamodbav:"auth_type"`
	Status        string                 `json:"status" dynamodbav:"status"`
	Version       int                    `json:"version" dynamodbav:"version"`

	// OAuth2 credentials
	AccessToken  string `json:"access_token,omitempty" dynamodbav:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty" dynamodbav:"refresh_token,omitempty"`
	TokenType    string `json:"token_type,omitempty" dynamodbav:"token_type,omitempty"`
	ExpiresAt    int64  `json:"expires_at,omitempty" dynamodbav:"expires_at,omitempty"`
	Scope        string `json:"scope,omitempty" dynamodbav:"scope,omitempty"`

	// API Key credentials
	APIKey    string `json:"api_key,omitempty" dynamodbav:"api_key,omitempty"`
	APISecret string `json:"api_secret,omitempty" dynamodbav:"api_secret,omitempty"`

	// Metadata & audit
	CredentialsMeta  map[string]interface{} `json:"credentials_meta,omitempty" dynamodbav:"credentials_meta,omitempty"`
	CreatedAt        int64                  `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt        int64                  `json:"updated_at" dynamodbav:"updated_at"`
	LastUsedAt       *int64                 `json:"last_used_at,omitempty" dynamodbav:"last_used_at,omitempty"`
	RevokedAt        *int64                 `json:"revoked_at,omitempty" dynamodbav:"revoked_at,omitempty"`

	// Token refresh tracking
	RefreshAttempts  int     `json:"refresh_attempts" dynamodbav:"refresh_attempts"`
	LastRefreshAt    *int64  `json:"last_refresh_at,omitempty" dynamodbav:"last_refresh_at,omitempty"`
	LastRefreshError *string `json:"last_refresh_error,omitempty" dynamodbav:"last_refresh_error,omitempty"`

	// TTL for automatic cleanup
	TTL *int64 `json:"ttl,omitempty" dynamodbav:"ttl,omitempty"`
}

// OAuthState represents a temporary OAuth state token for CSRF protection
type OAuthState struct {
	State      string                 `json:"state" dynamodbav:"state_id"`
	UserID     string                 `json:"user_id" dynamodbav:"user_id"`
	AccountID  string                 `json:"account_id" dynamodbav:"account_id"`
	PlatformID string                 `json:"platform_id" dynamodbav:"platform_id"`
	RedirectURI string                `json:"redirect_uri,omitempty" dynamodbav:"redirect_uri,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty" dynamodbav:"metadata,omitempty"`
	CreatedAt  int64                  `json:"created_at" dynamodbav:"created_at"`
	ExpiresAt  int64                  `json:"expires_at" dynamodbav:"expires_at"`
	TTL        int64                  `json:"ttl,omitempty" dynamodbav:"ttl,omitempty"`
}

// ========== HELPER TYPES ==========

// Helper represents a configured automation helper
type Helper struct {
	HelperID     string                 `json:"helper_id" dynamodbav:"helper_id"`
	AccountID    string                 `json:"account_id" dynamodbav:"account_id"`
	CreatedBy    string                 `json:"created_by" dynamodbav:"created_by"`
	ConnectionID string                 `json:"connection_id,omitempty" dynamodbav:"connection_id,omitempty"`
	ShortKey     string                 `json:"short_key" dynamodbav:"short_key"`
	Name         string                 `json:"name" dynamodbav:"name"`
	Description  string                 `json:"description" dynamodbav:"description"`
	HelperType   string                 `json:"helper_type" dynamodbav:"helper_type"`
	Category     string                 `json:"category" dynamodbav:"category"`
	Status       string                 `json:"status" dynamodbav:"status"`
	Config       map[string]interface{} `json:"config" dynamodbav:"config"`
	ConfigSchema map[string]interface{} `json:"config_schema,omitempty" dynamodbav:"config_schema,omitempty"`
	Enabled          bool                   `json:"enabled" dynamodbav:"enabled"`
	ExecutionCount   int64                  `json:"execution_count" dynamodbav:"execution_count"`
	LastExecutedAt   *time.Time             `json:"last_executed_at,omitempty" dynamodbav:"last_executed_at,omitempty"`
	ScheduleEnabled  bool                   `json:"schedule_enabled" dynamodbav:"schedule_enabled"`
	CronExpression   string                 `json:"cron_expression,omitempty" dynamodbav:"cron_expression,omitempty"`
	ScheduleRuleARN  string                 `json:"schedule_rule_arn,omitempty" dynamodbav:"schedule_rule_arn,omitempty"`
	LastScheduledAt  *time.Time             `json:"last_scheduled_at,omitempty" dynamodbav:"last_scheduled_at,omitempty"`
	NextScheduledAt  *time.Time             `json:"next_scheduled_at,omitempty" dynamodbav:"next_scheduled_at,omitempty"`
	CreatedAt        time.Time              `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at" dynamodbav:"updated_at"`
}

// HelperTemplate represents a predefined helper template from the library
type HelperTemplate struct {
	TemplateID   string                 `json:"template_id" dynamodbav:"template_id"`
	Name         string                 `json:"name" dynamodbav:"name"`
	Description  string                 `json:"description" dynamodbav:"description"`
	HelperType   string                 `json:"helper_type" dynamodbav:"helper_type"`
	Category     string                 `json:"category" dynamodbav:"category"`
	Icon         string                 `json:"icon" dynamodbav:"icon"`
	ConfigSchema map[string]interface{} `json:"config_schema" dynamodbav:"config_schema"`
	DefaultConfig map[string]interface{} `json:"default_config" dynamodbav:"default_config"`
	RequiresCRM  bool                   `json:"requires_crm" dynamodbav:"requires_crm"`
	SupportedCRMs []string              `json:"supported_crms" dynamodbav:"supported_crms"`
	Popularity   int                    `json:"popularity" dynamodbav:"popularity"`
	Status       string                 `json:"status" dynamodbav:"status"`
	CreatedAt    time.Time              `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" dynamodbav:"updated_at"`
}

// ========== EXECUTION TYPES ==========

// Execution represents a helper execution record
type Execution struct {
	ExecutionID  string                 `json:"execution_id" dynamodbav:"execution_id"`
	HelperID     string                 `json:"helper_id" dynamodbav:"helper_id"`
	HelperType   string                 `json:"helper_type,omitempty" dynamodbav:"helper_type,omitempty"`
	AccountID    string                 `json:"account_id" dynamodbav:"account_id"`
	UserID       string                 `json:"user_id,omitempty" dynamodbav:"user_id,omitempty"`
	APIKeyID     string                 `json:"api_key_id,omitempty" dynamodbav:"api_key_id,omitempty"`
	ConnectionID string                 `json:"connection_id,omitempty" dynamodbav:"connection_id,omitempty"`
	ContactID    string                 `json:"contact_id,omitempty" dynamodbav:"contact_id,omitempty"`
	Config       map[string]interface{} `json:"config,omitempty" dynamodbav:"config,omitempty"`
	Status       string                 `json:"status" dynamodbav:"status"`
	TriggerType  string                 `json:"trigger_type" dynamodbav:"trigger_type"`
	Input        map[string]interface{} `json:"input,omitempty" dynamodbav:"input,omitempty"`
	QueryParams  map[string]string      `json:"query_params,omitempty" dynamodbav:"query_params,omitempty"`
	Output       map[string]interface{} `json:"output,omitempty" dynamodbav:"output,omitempty"`
	ErrorMessage string                 `json:"error_message,omitempty" dynamodbav:"error_message,omitempty"`
	DurationMs   int64                  `json:"duration_ms" dynamodbav:"duration_ms"`
	CreatedAt    string                 `json:"created_at" dynamodbav:"created_at"`
	StartedAt    time.Time              `json:"started_at" dynamodbav:"started_at"`
	CompletedAt          *time.Time             `json:"completed_at,omitempty" dynamodbav:"completed_at,omitempty"`
	TTL                  *int64                 `json:"ttl,omitempty" dynamodbav:"ttl,omitempty"`
	StripeReported       bool                   `json:"stripe_reported,omitempty" dynamodbav:"stripe_reported,omitempty"`
	StripeUsageRecordID  string                 `json:"stripe_usage_record_id,omitempty" dynamodbav:"stripe_usage_record_id,omitempty"`
}

// ========== API KEY TYPES ==========

// APIKey represents an API key for external helper execution
type APIKey struct {
	KeyID       string     `json:"key_id" dynamodbav:"key_id"`
	AccountID   string     `json:"account_id" dynamodbav:"account_id"`
	CreatedBy   string     `json:"created_by" dynamodbav:"created_by"`
	Name        string     `json:"name" dynamodbav:"name"`
	KeyHash     string     `json:"key_hash" dynamodbav:"key_hash"`
	KeyPrefix   string     `json:"key_prefix" dynamodbav:"key_prefix"`
	Permissions []string   `json:"permissions" dynamodbav:"permissions"`
	Status      string     `json:"status" dynamodbav:"status"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty" dynamodbav:"last_used_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at" dynamodbav:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" dynamodbav:"expires_at,omitempty"`
}

// ========== CHAT TYPES ==========

// ChatConversation represents a chat conversation in the system
type ChatConversation struct {
	ConversationID string  `json:"conversation_id" dynamodbav:"conversation_id"`
	AccountID      string  `json:"account_id" dynamodbav:"account_id"`
	UserID         string  `json:"user_id" dynamodbav:"user_id"`
	Title          string  `json:"title" dynamodbav:"title"`
	CreatedAt      string  `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt      string  `json:"updated_at" dynamodbav:"updated_at"`
	DeletedAt      *string `json:"deleted_at,omitempty" dynamodbav:"deleted_at,omitempty"`
	MessageCount   int     `json:"message_count" dynamodbav:"message_count"`
	TTL            *int64  `json:"ttl,omitempty" dynamodbav:"ttl,omitempty"` // 90 days auto-cleanup
}

// ChatMessage represents a message in a chat conversation
type ChatMessage struct {
	MessageID      string                 `json:"message_id" dynamodbav:"message_id"`
	ConversationID string                 `json:"conversation_id" dynamodbav:"conversation_id"`
	Sequence       int                    `json:"sequence" dynamodbav:"sequence"`
	Role           string                 `json:"role" dynamodbav:"role"` // "user" or "assistant"
	Content        string                 `json:"content" dynamodbav:"content"`
	ToolCalls      []ToolCall             `json:"tool_calls,omitempty" dynamodbav:"tool_calls,omitempty"`
	ToolResults    []ToolResult           `json:"tool_results,omitempty" dynamodbav:"tool_results,omitempty"`
	CreatedAt      string                 `json:"created_at" dynamodbav:"created_at"`
	Metadata       map[string]interface{} `json:"metadata,omitempty" dynamodbav:"metadata,omitempty"`
}

// ToolCall represents a tool/function call in a message
// This is the format used by MCP service and compatible with Groq API
type ToolCall struct {
	ID       string       `json:"id" dynamodbav:"id"`
	Type     string       `json:"type" dynamodbav:"type"` // "function"
	Function FunctionCall `json:"function" dynamodbav:"function"`
}

// FunctionCall represents the function details in a tool call
type FunctionCall struct {
	Name      string `json:"name" dynamodbav:"name"`
	Arguments string `json:"arguments" dynamodbav:"arguments"` // JSON string
}

// ToolResult represents the result of a tool/function call
type ToolResult struct {
	ToolCallID string `json:"tool_call_id" dynamodbav:"tool_call_id"`
	Result     string `json:"result" dynamodbav:"result"`
}

// ========== PLAN TYPES ==========

// PlanLimits defines limits for an account plan
type PlanLimits struct {
	MaxHelpers      int `json:"max_helpers"`
	MaxExecutions   int `json:"max_executions"`
	MaxConnections  int `json:"max_connections"`
	MaxTeamMembers  int `json:"max_team_members"`
	MaxAPIKeys      int `json:"max_api_keys"`
}

// ========== GROQ API TYPES ==========

// GroqChatRequest represents a Groq Chat API request
type GroqChatRequest struct {
	Model       string        `json:"model"`
	Messages    []GroqMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Tools       []GroqTool    `json:"tools,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
}

// GroqMessage represents a message in the Groq format
type GroqMessage struct {
	Role       string         `json:"role"` // "system", "user", "assistant", "tool"
	Content    string         `json:"content,omitempty"`
	ToolCalls  []GroqToolCall `json:"tool_calls,omitempty"`
	ToolCallID string         `json:"tool_call_id,omitempty"` // for role="tool"
	Name       string         `json:"name,omitempty"`         // for role="tool"
}

// GroqToolCall represents a tool call in Groq format
type GroqToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function GroqFunctionCall `json:"function"`
}

// GroqFunctionCall represents a function call in Groq format
type GroqFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// GroqTool represents a tool definition for Groq
type GroqTool struct {
	Type     string          `json:"type"`
	Function GroqFunctionDef `json:"function"`
}

// GroqFunctionDef represents a function definition for Groq
type GroqFunctionDef struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// GroqChatResponse represents a Groq Chat API response
type GroqChatResponse struct {
	ID      string       `json:"id"`
	Object  string       `json:"object"`
	Created int64        `json:"created"`
	Model   string       `json:"model"`
	Choices []GroqChoice `json:"choices"`
	Usage   GroqUsage    `json:"usage"`
}

// GroqChoice represents a choice in the Groq response
type GroqChoice struct {
	Index        int         `json:"index"`
	Message      GroqMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

// GroqUsage represents token usage information
type GroqUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// GroqStreamChunk represents a chunk in a streaming response
type GroqStreamChunk struct {
	ID      string            `json:"id"`
	Object  string            `json:"object"`
	Created int64             `json:"created"`
	Model   string            `json:"model"`
	Choices []GroqStreamDelta `json:"choices"`
}

// GroqStreamDelta represents a delta in a streaming chunk
type GroqStreamDelta struct {
	Index        int       `json:"index"`
	Delta        GroqDelta `json:"delta"`
	FinishReason *string   `json:"finish_reason"`
}

// GroqDelta represents delta content in streaming
type GroqDelta struct {
	Role      string         `json:"role,omitempty"`
	Content   string         `json:"content,omitempty"`
	ToolCalls []GroqToolCall `json:"tool_calls,omitempty"`
}

// ========== CHAT REQUEST/RESPONSE TYPES ==========

// CreateConversationRequest represents a request to create a new conversation
type CreateConversationRequest struct {
	Title string `json:"title,omitempty"`
}

// CreateConversationResponse represents the response for conversation creation
type CreateConversationResponse struct {
	ConversationID string `json:"conversation_id"`
	Title          string `json:"title"`
	Status         string `json:"status"`
	CreatedAt      string `json:"created_at"`
}

// ListConversationsResponse represents the response for listing conversations
type ListConversationsResponse struct {
	Conversations []ChatConversation `json:"conversations"`
}

// GetConversationResponse represents the response for getting a conversation with messages
type GetConversationResponse struct {
	Conversation ChatConversation `json:"conversation"`
	Messages     []ChatMessage    `json:"messages"`
}

// SendMessageRequest represents a request to send a message
type SendMessageRequest struct {
	Content string `json:"content"`
}

// StreamChatResponse represents a chunk of streaming chat response
type StreamChatResponse struct {
	Type       string      `json:"type"` // "content", "tool_call", "tool_result", "done"
	Content    string      `json:"content,omitempty"`
	ToolCall   *ToolCall   `json:"tool_call,omitempty"`
	ToolResult *ToolResult `json:"tool_result,omitempty"`
	Done       bool        `json:"done,omitempty"`
}

// ========== EMAIL TYPES ==========

// EmailTemplate represents an email template stored in DynamoDB
type EmailTemplate struct {
	TemplateID   string          `json:"template_id" dynamodbav:"template_id"`
	AccountID    string          `json:"account_id,omitempty" dynamodbav:"account_id,omitempty"` // Empty for system templates
	Name         string          `json:"name" dynamodbav:"name"`
	Subject      string          `json:"subject" dynamodbav:"subject"`
	HTMLTemplate string          `json:"html_template" dynamodbav:"html_template"`
	TextTemplate string          `json:"text_template" dynamodbav:"text_template"`
	Variables    []EmailVariable `json:"variables,omitempty" dynamodbav:"variables,omitempty"`
	IsSystem     bool            `json:"is_system" dynamodbav:"is_system"`
	IsActive     bool            `json:"is_active" dynamodbav:"is_active"`
	CreatedAt    string          `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt    string          `json:"updated_at" dynamodbav:"updated_at"`
}

// EmailVariable represents a template variable
type EmailVariable struct {
	Name        string `json:"name" dynamodbav:"name"`
	Description string `json:"description" dynamodbav:"description"`
	Required    bool   `json:"required" dynamodbav:"required"`
}

// EmailLog represents a sent email record
type EmailLog struct {
	EmailID        string `json:"email_id" dynamodbav:"email_id"`
	AccountID      string `json:"account_id" dynamodbav:"account_id"`
	RecipientEmail string `json:"recipient_email" dynamodbav:"recipient_email"`
	Subject        string `json:"subject" dynamodbav:"subject"`
	TemplateID     string `json:"template_id,omitempty" dynamodbav:"template_id,omitempty"`
	Status         string `json:"status" dynamodbav:"status"` // sent, failed, bounced
	MessageID      string `json:"message_id,omitempty" dynamodbav:"message_id,omitempty"`
	ErrorMessage   string `json:"error_message,omitempty" dynamodbav:"error_message,omitempty"`
	CreatedAt      string `json:"created_at" dynamodbav:"created_at"`
	SentAt         string `json:"sent_at,omitempty" dynamodbav:"sent_at,omitempty"`
	TTL            int64  `json:"ttl" dynamodbav:"ttl"` // Auto-delete after 90 days
}

// EmailVerification represents an email verification record
type EmailVerification struct {
	VerificationID string `json:"verification_id" dynamodbav:"verification_id"`
	Email          string `json:"email" dynamodbav:"email"`
	Token          string `json:"token" dynamodbav:"token"`
	ExpiresAt      int64  `json:"expires_at" dynamodbav:"expires_at"` // Unix timestamp, also used for TTL
	CreatedAt      string `json:"created_at" dynamodbav:"created_at"`
	VerifiedAt     string `json:"verified_at,omitempty" dynamodbav:"verified_at,omitempty"`
	Status         string `json:"status" dynamodbav:"status"` // pending, verified, expired
}

// SendEmailRequest represents a request to send an email
type SendEmailRequest struct {
	To           []string               `json:"to"`
	Cc           []string               `json:"cc,omitempty"`
	Bcc          []string               `json:"bcc,omitempty"`
	Subject      string                 `json:"subject"`
	HTMLBody     string                 `json:"html_body,omitempty"`
	TextBody     string                 `json:"text_body,omitempty"`
	TemplateID   string                 `json:"template_id,omitempty"`
	TemplateData map[string]interface{} `json:"template_data,omitempty"`
	FromEmail    string                 `json:"from_email,omitempty"`
	FromName     string                 `json:"from_name,omitempty"`
	ReplyTo      []string               `json:"reply_to,omitempty"`
}

// SendEmailResponse represents the response for sending an email
type SendEmailResponse struct {
	EmailID   string `json:"email_id"`
	MessageID string `json:"message_id"`
	Status    string `json:"status"`
}

// GetEmailHistoryResponse represents the response for getting email history
type GetEmailHistoryResponse struct {
	Emails []EmailLog `json:"emails"`
	Total  int        `json:"total"`
}

// CreateTemplateRequest represents a request to create an email template
type CreateTemplateRequest struct {
	Name         string          `json:"name"`
	Subject      string          `json:"subject"`
	HTMLTemplate string          `json:"html_template"`
	TextTemplate string          `json:"text_template,omitempty"`
	Variables    []EmailVariable `json:"variables,omitempty"`
}

// UpdateTemplateRequest represents a request to update an email template
type UpdateTemplateRequest struct {
	Name         string          `json:"name,omitempty"`
	Subject      string          `json:"subject,omitempty"`
	HTMLTemplate string          `json:"html_template,omitempty"`
	TextTemplate string          `json:"text_template,omitempty"`
	Variables    []EmailVariable `json:"variables,omitempty"`
	IsActive     *bool           `json:"is_active,omitempty"`
}

// ListTemplatesResponse represents the response for listing templates
type ListTemplatesResponse struct {
	Templates []EmailTemplate `json:"templates"`
	Total     int             `json:"total"`
}
