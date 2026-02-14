package billing

// PlanConfig defines the limits and pricing for a subscription plan.
type PlanConfig struct {
	Name           string
	MaxHelpers     int
	MaxConnections int
	MaxAPIKeys     int
	MaxTeamMembers int
	MaxExecutions  int     // included per billing cycle
	PriceMonthly   int     // in USD
	OverageRate    float64 // cost per execution beyond included (0 = blocked at limit)
}

// Plans defines all available subscription tiers.
// "free" is the sandbox tier for users who haven't selected a paid plan.
var Plans = map[string]PlanConfig{
	"free": {
		Name:           "Sandbox",
		MaxHelpers:     1,
		MaxConnections: 1,
		MaxAPIKeys:     1,
		MaxTeamMembers: 1,
		MaxExecutions:  100,
		PriceMonthly:   0,
		OverageRate:    0, // hard-blocked at limit
	},
	"start": {
		Name:           "Start",
		MaxHelpers:     10,
		MaxConnections: 2,
		MaxAPIKeys:     2,
		MaxTeamMembers: 2,
		MaxExecutions:  5000,
		PriceMonthly:   39,
		OverageRate:    0.01,
	},
	"grow": {
		Name:           "Grow",
		MaxHelpers:     50,
		MaxConnections: 5,
		MaxAPIKeys:     10,
		MaxTeamMembers: 5,
		MaxExecutions:  25000,
		PriceMonthly:   59,
		OverageRate:    0.008,
	},
	"deliver": {
		Name:           "Deliver",
		MaxHelpers:     999999,
		MaxConnections: 20,
		MaxAPIKeys:     100,
		MaxTeamMembers: 100,
		MaxExecutions:  100000,
		PriceMonthly:   79,
		OverageRate:    0.005,
	},
}

// GetPlan returns the PlanConfig for the given plan name.
// Returns the sandbox (free) config if the plan is unknown.
func GetPlan(plan string) PlanConfig {
	if cfg, ok := Plans[plan]; ok {
		return cfg
	}
	return Plans["free"]
}

// GetPlanLabel returns the display name for a plan.
func GetPlanLabel(plan string) string {
	if cfg, ok := Plans[plan]; ok {
		return cfg.Name
	}
	return "Sandbox"
}

// IsPaidPlan returns true if the plan is a paid subscription tier.
func IsPaidPlan(plan string) bool {
	return plan == "start" || plan == "grow" || plan == "deliver"
}
