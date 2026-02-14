package email

import (
	"fmt"
	"regexp"
	"strings"
)

// EmailValidator provides advanced email validation
type EmailValidator struct {
	disposableDomains map[string]bool
}

// ValidationResult contains the result of email validation
type ValidationResult struct {
	IsValid    bool
	Email      string
	Error      string
	RiskScore  int    // 0-100, higher = riskier
	RiskReason string // Why this email is risky
}

// NewEmailValidator creates a new email validator
func NewEmailValidator() *EmailValidator {
	return &EmailValidator{
		disposableDomains: getDisposableDomains(),
	}
}

// ValidateEmail performs comprehensive email validation
func (v *EmailValidator) ValidateEmail(email string) *ValidationResult {
	email = strings.TrimSpace(strings.ToLower(email))

	result := &ValidationResult{
		Email:     email,
		RiskScore: 0,
	}

	// Basic format validation
	if !v.isValidFormat(email) {
		result.IsValid = false
		result.Error = "invalid email format"
		result.RiskScore = 100
		return result
	}

	// Extract domain
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		result.IsValid = false
		result.Error = "invalid email structure"
		result.RiskScore = 100
		return result
	}

	domain := parts[1]
	localPart := parts[0]

	// Check for disposable email domains
	if v.isDisposableDomain(domain) {
		result.IsValid = false
		result.Error = "disposable email domain not allowed"
		result.RiskScore = 90
		result.RiskReason = "disposable domain"
		return result
	}

	// Check for suspicious patterns
	riskScore := 0
	riskReasons := []string{}

	// Check for role-based emails (noreply, admin, info, etc.)
	if v.isRoleBasedEmail(localPart) {
		riskScore += 20
		riskReasons = append(riskReasons, "role-based email")
	}

	// Check for suspicious characters or patterns
	if v.hasSuspiciousPattern(localPart) {
		riskScore += 30
		riskReasons = append(riskReasons, "suspicious pattern")
	}

	// Check for excessive length
	if len(email) > 254 || len(localPart) > 64 || len(domain) > 253 {
		result.IsValid = false
		result.Error = "email exceeds maximum length"
		result.RiskScore = 80
		return result
	}

	result.IsValid = true
	result.RiskScore = riskScore
	if len(riskReasons) > 0 {
		result.RiskReason = strings.Join(riskReasons, ", ")
	}

	return result
}

// ValidateBatch validates multiple email addresses
func (v *EmailValidator) ValidateBatch(emails []string) []*ValidationResult {
	results := make([]*ValidationResult, len(emails))
	for i, email := range emails {
		results[i] = v.ValidateEmail(email)
	}
	return results
}

// isValidFormat checks basic email format using regex
func (v *EmailValidator) isValidFormat(email string) bool {
	// RFC 5322 simplified regex pattern
	pattern := `^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`
	matched, err := regexp.MatchString(pattern, email)
	return err == nil && matched
}

// isDisposableDomain checks if the domain is a known disposable email provider
func (v *EmailValidator) isDisposableDomain(domain string) bool {
	_, exists := v.disposableDomains[domain]
	return exists
}

// isRoleBasedEmail checks if the local part is a role-based address
func (v *EmailValidator) isRoleBasedEmail(localPart string) bool {
	roleBased := map[string]bool{
		"admin":     true,
		"noreply":   true,
		"no-reply":  true,
		"info":      true,
		"support":   true,
		"help":      true,
		"sales":     true,
		"marketing": true,
		"contact":   true,
		"postmaster": true,
		"abuse":     true,
		"webmaster": true,
		"hostmaster": true,
		"root":      true,
	}

	return roleBased[localPart]
}

// hasSuspiciousPattern checks for suspicious patterns in email
func (v *EmailValidator) hasSuspiciousPattern(localPart string) bool {
	// Check for consecutive dots
	if strings.Contains(localPart, "..") {
		return true
	}

	// Check for leading or trailing dots
	if strings.HasPrefix(localPart, ".") || strings.HasSuffix(localPart, ".") {
		return true
	}

	// Check for excessive plus signs (common in spam)
	plusCount := strings.Count(localPart, "+")
	if plusCount > 1 {
		return true
	}

	// Check for excessive numbers (often indicates auto-generated)
	digitCount := 0
	for _, char := range localPart {
		if char >= '0' && char <= '9' {
			digitCount++
		}
	}
	if digitCount > len(localPart)/2 {
		return true
	}

	return false
}

// SanitizeEmail normalizes and cleans an email address
func SanitizeEmail(email string) string {
	email = strings.TrimSpace(email)
	email = strings.ToLower(email)
	return email
}

// ExtractDomain extracts the domain from an email address
func ExtractDomain(email string) (string, error) {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid email format")
	}
	return parts[1], nil
}

// getDisposableDomains returns a map of known disposable email domains
func getDisposableDomains() map[string]bool {
	domains := []string{
		// Common disposable email providers
		"10minutemail.com",
		"guerrillamail.com",
		"mailinator.com",
		"tempmail.com",
		"throwaway.email",
		"yopmail.com",
		"temp-mail.org",
		"fakeinbox.com",
		"trashmail.com",
		"getnada.com",
		"dispostable.com",
		"sharklasers.com",
		"guerrillamailblock.com",
		"spam4.me",
		"grr.la",
		"maildrop.cc",
		"mintemail.com",
		"mytemp.email",
		"tempinbox.com",
		"trbvm.com",
		"mohmal.com",
		"emailondeck.com",
		"throwawaymail.com",
		"spamgourmet.com",
		"mailnesia.com",
		"deadaddress.com",
		"anonbox.net",
		"binkmail.com",
		"bobmail.info",
		"deadfake.cf",
		"discard.email",
		"emailsensei.com",
		"eyepaste.com",
		"getairmail.com",
		"hidemail.de",
		"incognitomail.com",
		"jetable.org",
		"meltmail.com",
		"mytrashmail.com",
		"no-spam.ws",
		"noclickemail.com",
		"pookmail.com",
		"quickinbox.com",
		"recode.me",
		"safetymail.info",
		"spambox.us",
		"spamfree24.org",
		"tempr.email",
		"throam.com",
		"tmail.ws",
		"tmailinator.com",
		"trashmailer.com",
		"wegwerfmail.de",
		"zoemail.net",
	}

	domainMap := make(map[string]bool, len(domains))
	for _, domain := range domains {
		domainMap[domain] = true
	}
	return domainMap
}
