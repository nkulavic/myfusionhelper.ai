package translate

// FieldMapper translates standardized field names to CRM-specific field keys.
// This is a static, in-memory mapping requiring no API calls.
type FieldMapper struct {
	platformSlug  string
	standardToCRM map[string]string // "first_name" -> "given_name" (Keap)
	crmToStandard map[string]string // "given_name" -> "first_name" (reverse)
}

// platformFieldMaps defines the standardized-to-CRM field mapping for each platform.
// The standardized names are snake_case and match NormalizedContact field conventions.
var platformFieldMaps = map[string]map[string]string{
	"keap": {
		"first_name":  "given_name",
		"last_name":   "family_name",
		"email":       "email",
		"phone":       "phone",
		"company":     "company",
		"job_title":   "job_title",
		"website":     "website",
		"birthday":    "birthday",
		"anniversary": "anniversary",
		"address1":    "line1",
		"address2":    "line2",
		"city":        "locality",
		"state":       "region",
		"zip":         "postal_code",
		"country":     "country_code",
	},
	"gohighlevel": {
		"first_name": "firstName",
		"last_name":  "lastName",
		"email":      "email",
		"phone":      "phone",
		"company":    "companyName",
		"job_title":  "jobTitle",
		"website":    "website",
		"birthday":   "dateOfBirth",
		"address1":   "address1",
		"city":       "city",
		"state":      "state",
		"zip":        "postalCode",
		"country":    "country",
		"source":     "source",
		"timezone":   "timezone",
	},
	"activecampaign": {
		"first_name": "firstName",
		"last_name":  "lastName",
		"email":      "email",
		"phone":      "phone",
		"company":    "orgname",
		"job_title":  "jobTitle",
	},
	"ontraport": {
		"first_name": "firstname",
		"last_name":  "lastname",
		"email":      "email",
		"phone":      "office_phone",
		"company":    "company",
		"job_title":  "title",
		"website":    "website",
		"birthday":   "birthday",
		"address1":   "address",
		"address2":   "address2",
		"city":       "city",
		"state":      "state",
		"zip":        "zip",
		"country":    "country",
		"sms_number": "sms_number",
		"fax":        "fax",
	},
}

// NewFieldMapper creates a FieldMapper for the given platform.
func NewFieldMapper(platformSlug string) *FieldMapper {
	fm := &FieldMapper{
		platformSlug:  platformSlug,
		standardToCRM: make(map[string]string),
		crmToStandard: make(map[string]string),
	}

	if mapping, ok := platformFieldMaps[platformSlug]; ok {
		for standard, crm := range mapping {
			fm.standardToCRM[standard] = crm
			fm.crmToStandard[crm] = standard
		}
	}

	return fm
}

// Resolve translates a field key to the CRM-specific equivalent.
// Resolution order:
//  1. If the key is a known standardized name, return the CRM-specific key.
//  2. If the key is already a CRM-native key, return it unchanged.
//  3. Otherwise, return the key unchanged (it may be a custom field ID or label).
func (fm *FieldMapper) Resolve(fieldKey string) string {
	// Check if it's a standardized name
	if crmKey, ok := fm.standardToCRM[fieldKey]; ok {
		return crmKey
	}

	// Already a CRM-native key or an unknown key — pass through
	return fieldKey
}

// IsStandardField returns true if the given key is a known standardized field name.
func (fm *FieldMapper) IsStandardField(fieldKey string) bool {
	_, ok := fm.standardToCRM[fieldKey]
	return ok
}

// IsCRMNativeKey returns true if the given key is a known CRM-native field key.
func (fm *FieldMapper) IsCRMNativeKey(fieldKey string) bool {
	_, ok := fm.crmToStandard[fieldKey]
	return ok
}

// ToStandard translates a CRM-native key back to the standardized name.
// Returns the original key if no mapping exists.
func (fm *FieldMapper) ToStandard(crmKey string) string {
	if standard, ok := fm.crmToStandard[crmKey]; ok {
		return standard
	}
	return crmKey
}
