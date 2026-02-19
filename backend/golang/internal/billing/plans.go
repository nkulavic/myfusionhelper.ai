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
	PriceAnnually  int     // in USD (total per year)
	OverageRate    float64 // cost per execution beyond included (0 = blocked at limit)
}

// Plans defines all available subscription tiers.
// "free" is the legacy sandbox tier (kept for backward compat).
// "trial" is the 14-day free trial with Start-level limits.
var Plans = map[string]PlanConfig{
	"free": {
		Name:           "Sandbox",
		MaxHelpers:     1,
		MaxConnections: 1,
		MaxAPIKeys:     1,
		MaxTeamMembers: 1,
		MaxExecutions:  100,
		PriceMonthly:   0,
		PriceAnnually:  0,
		OverageRate:    0, // hard-blocked at limit
	},
	"trial": {
		Name:           "Free Trial",
		MaxHelpers:     10,
		MaxConnections: 2,
		MaxAPIKeys:     2,
		MaxTeamMembers: 2,
		MaxExecutions:  5000,
		PriceMonthly:   0,
		PriceAnnually:  0,
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
		PriceAnnually:  396,
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
		PriceAnnually:  588,
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
		PriceAnnually:  792,
		OverageRate:    0.005,
	},
}

// GetPlan returns the PlanConfig for the given plan name.
// Returns the trial config if the plan is unknown.
func GetPlan(plan string) PlanConfig {
	if cfg, ok := Plans[plan]; ok {
		return cfg
	}
	return Plans["trial"]
}

// GetPlanLabel returns the display name for a plan.
func GetPlanLabel(plan string) string {
	if cfg, ok := Plans[plan]; ok {
		return cfg.Name
	}
	return "Free Trial"
}

// IsPaidPlan returns true if the plan is a paid subscription tier.
func IsPaidPlan(plan string) bool {
	return plan == "start" || plan == "grow" || plan == "deliver"
}

// IsTrialPlan returns true if the plan is a free trial or legacy free tier.
func IsTrialPlan(plan string) bool {
	return plan == "trial" || plan == "free"
}
