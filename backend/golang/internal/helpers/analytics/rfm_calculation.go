package analytics

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/myfusionhelper/api/internal/helpers"
)

// NewRFMCalculation creates a new RFMCalculation helper instance
func NewRFMCalculation() helpers.Helper { return &RFMCalculation{} }

func init() {
	helpers.Register("rfm_calculation", func() helpers.Helper { return &RFMCalculation{} })
}

// RFMCalculation computes Recency, Frequency, and Monetary scores for a contact
// based on their invoice/order history and configurable thresholds.
// Ported from legacy PHP rfm_calculation helper.
type RFMCalculation struct{}

func (h *RFMCalculation) GetName() string     { return "RFM Calculation" }
func (h *RFMCalculation) GetType() string     { return "rfm_calculation" }
func (h *RFMCalculation) GetCategory() string { return "analytics" }
func (h *RFMCalculation) GetDescription() string {
	return "Calculate Recency, Frequency, and Monetary (RFM) customer metrics from order history"
}
func (h *RFMCalculation) RequiresCRM() bool       { return true }
func (h *RFMCalculation) SupportedCRMs() []string { return nil }

func (h *RFMCalculation) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"options": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"score_range": map[string]interface{}{
						"type":        "object",
						"description": "Configure score range (default: 1-5, allows 1-10)",
						"properties": map[string]interface{}{
							"min": map[string]interface{}{"type": "number", "description": "Minimum score value", "default": 1},
							"max": map[string]interface{}{"type": "number", "description": "Maximum score value", "default": 5},
						},
					},
					"percentile_based_scoring": map[string]interface{}{
						"type":        "boolean",
						"description": "Use percentile-based scoring instead of fixed thresholds",
						"default":     false,
					},
					"recency_calculation": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"1_2_threshold": map[string]interface{}{"type": "number", "description": "Days threshold for score 2 (e.g. 120)"},
							"2_3_threshold": map[string]interface{}{"type": "number", "description": "Days threshold for score 3 (e.g. 90)"},
							"3_4_threshold": map[string]interface{}{"type": "number", "description": "Days threshold for score 4 (e.g. 60)"},
							"4_5_threshold": map[string]interface{}{"type": "number", "description": "Days threshold for score 5 (e.g. 30)"},
						},
					},
					"frequency_calculation": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"1_2_threshold": map[string]interface{}{"type": "number", "description": "Order count threshold for score 2"},
							"2_3_threshold": map[string]interface{}{"type": "number", "description": "Order count threshold for score 3"},
							"3_4_threshold": map[string]interface{}{"type": "number", "description": "Order count threshold for score 4"},
							"4_5_threshold": map[string]interface{}{"type": "number", "description": "Order count threshold for score 5"},
						},
					},
					"monetary_calculation": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"1_2_threshold": map[string]interface{}{"type": "number", "description": "Total spend threshold for score 2"},
							"2_3_threshold": map[string]interface{}{"type": "number", "description": "Total spend threshold for score 3"},
							"3_4_threshold": map[string]interface{}{"type": "number", "description": "Total spend threshold for score 4"},
							"4_5_threshold": map[string]interface{}{"type": "number", "description": "Total spend threshold for score 5"},
						},
					},
					"segment_labels": map[string]interface{}{
						"type":        "boolean",
						"description": "Enable RFM segment labels (Champions, Loyal, At Risk, Hibernating, Lost)",
						"default":     false,
					},
					"segment_tags": map[string]interface{}{
						"type":        "object",
						"description": "Tag IDs to apply for each RFM segment",
						"properties": map[string]interface{}{
							"champions_tag":    map[string]interface{}{"type": "string", "description": "Tag for Champions (555)"},
							"loyal_tag":        map[string]interface{}{"type": "string", "description": "Tag for Loyal Customers (4-5, 4-5, 4-5)"},
							"at_risk_tag":      map[string]interface{}{"type": "string", "description": "Tag for At Risk (2-3, 2-3, 4-5)"},
							"hibernating_tag":  map[string]interface{}{"type": "string", "description": "Tag for Hibernating (1-2, 1-2, 2-3)"},
							"lost_tag":         map[string]interface{}{"type": "string", "description": "Tag for Lost (1, 1, 1-2)"},
							"promising_tag":    map[string]interface{}{"type": "string", "description": "Tag for Promising (3-4, 1-2, 1-2)"},
							"need_attention_tag": map[string]interface{}{"type": "string", "description": "Tag for Need Attention (2-3, 2-3, 2-3)"},
						},
					},
					"save_data": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"recency_score":         map[string]interface{}{"type": "string"},
							"frequency_score":       map[string]interface{}{"type": "string"},
							"monetary_score":        map[string]interface{}{"type": "string"},
							"rfm_composite_score":   map[string]interface{}{"type": "string"},
							"rfm_segment_label":     map[string]interface{}{"type": "string", "description": "Field to store segment label"},
							"total_order_value":     map[string]interface{}{"type": "string"},
							"average_order_value":   map[string]interface{}{"type": "string"},
							"total_order_count":     map[string]interface{}{"type": "string"},
							"first_order_date":      map[string]interface{}{"type": "string"},
							"first_order_value":     map[string]interface{}{"type": "string"},
							"last_order_date":       map[string]interface{}{"type": "string"},
							"last_order_value":      map[string]interface{}{"type": "string"},
							"days_since_last_order": map[string]interface{}{"type": "string"},
						},
						"description": "Map of metric names to contact field keys for saving results",
					},
				},
				"description": "RFM calculation options including thresholds and save field mappings",
			},
		},
		"required": []string{"options"},
	}
}

func (h *RFMCalculation) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["options"]; !ok {
		return fmt.Errorf("options is required")
	}
	return nil
}

func (h *RFMCalculation) Execute(ctx context.Context, input helpers.HelperInput) (*helpers.HelperOutput, error) {
	options := extractMap(input.Config["options"])

	output := &helpers.HelperOutput{
		Actions: make([]helpers.HelperAction, 0),
		Logs:    make([]string, 0),
	}

	// Parse score range configuration
	scoreMin := 1
	scoreMax := 5
	if scoreRange, ok := options["score_range"].(map[string]interface{}); ok {
		if minVal, ok := scoreRange["min"].(float64); ok && minVal > 0 {
			scoreMin = int(minVal)
		}
		if maxVal, ok := scoreRange["max"].(float64); ok && maxVal > float64(scoreMin) {
			scoreMax = int(maxVal)
		}
	}

	percentileBasedScoring := false
	if pbs, ok := options["percentile_based_scoring"].(bool); ok {
		percentileBasedScoring = pbs
	}

	segmentLabelsEnabled := false
	if sle, ok := options["segment_labels"].(bool); ok {
		segmentLabelsEnabled = sle
	}

	// Get order/invoice data from the CRM
	// Query for total spend, total orders, etc. via composite keys
	totalSpendVal, _ := input.Connector.GetContactFieldValue(ctx, input.ContactID, "_related.invoice.sum.TotalPaid")
	totalOrdersVal, _ := input.Connector.GetContactFieldValue(ctx, input.ContactID, "_related.invoice.count")
	lastOrderDateVal, _ := input.Connector.GetContactFieldValue(ctx, input.ContactID, "_related.invoice.last.DateCreated")
	lastOrderValueVal, _ := input.Connector.GetContactFieldValue(ctx, input.ContactID, "_related.invoice.last.InvoiceTotal")
	firstOrderDateVal, _ := input.Connector.GetContactFieldValue(ctx, input.ContactID, "_related.invoice.first.DateCreated")
	firstOrderValueVal, _ := input.Connector.GetContactFieldValue(ctx, input.ContactID, "_related.invoice.first.InvoiceTotal")
	totalInvoiceVal, _ := input.Connector.GetContactFieldValue(ctx, input.ContactID, "_related.invoice.sum.InvoiceTotal")

	totalSpend := toFloat64(totalSpendVal)
	totalOrders := toFloat64(totalOrdersVal)
	lastOrderDate := fmt.Sprintf("%v", lastOrderDateVal)
	lastOrderValue := toFloat64(lastOrderValueVal)
	firstOrderDate := fmt.Sprintf("%v", firstOrderDateVal)
	firstOrderValue := toFloat64(firstOrderValueVal)
	totalInvoiceAmount := toFloat64(totalInvoiceVal)

	if totalOrders == 0 {
		output.Success = true
		output.Message = "No orders found for contact"
		output.Logs = append(output.Logs, output.Message)
		return output, nil
	}

	averageOrderSize := math.Round(totalInvoiceAmount / totalOrders)
	actualPaymentPerOrder := math.Round(totalSpend / totalOrders)
	totalSpendRounded := math.Round(totalSpend)
	totalInvoiceRounded := math.Round(totalInvoiceAmount)

	// Calculate days since last order
	daysSinceLastOrder := 0
	if lastOrderDate != "" && lastOrderDate != "<nil>" {
		parsedDate, err := time.Parse("2006-01-02", lastOrderDate[:minInt(10, len(lastOrderDate))])
		if err == nil {
			daysSinceLastOrder = int(time.Since(parsedDate).Hours() / 24)
		}
	}

	// Calculate Recency score
	recencyCalc := extractMap(options["recency_calculation"])
	recencyScore := scoreMin
	if percentileBasedScoring {
		// Simplified percentile: lower days = higher score
		recencyScore = calculatePercentileScore(float64(daysSinceLastOrder), 0, 365, scoreMin, scoreMax, true)
	} else {
		// Fixed thresholds
		if daysSinceLastOrder <= int(toFloat64(recencyCalc["1_2_threshold"])) {
			recencyScore = scaleScore(2, scoreMin, scoreMax)
		}
		if daysSinceLastOrder <= int(toFloat64(recencyCalc["2_3_threshold"])) {
			recencyScore = scaleScore(3, scoreMin, scoreMax)
		}
		if daysSinceLastOrder <= int(toFloat64(recencyCalc["3_4_threshold"])) {
			recencyScore = scaleScore(4, scoreMin, scoreMax)
		}
		if daysSinceLastOrder <= int(toFloat64(recencyCalc["4_5_threshold"])) {
			recencyScore = scaleScore(5, scoreMin, scoreMax)
		}
	}

	// Calculate Frequency score
	frequencyCalc := extractMap(options["frequency_calculation"])
	frequencyScore := scoreMin
	if percentileBasedScoring {
		frequencyScore = calculatePercentileScore(totalOrders, 0, 100, scoreMin, scoreMax, false)
	} else {
		// Fixed thresholds
		if totalOrders >= toFloat64(frequencyCalc["1_2_threshold"]) {
			frequencyScore = scaleScore(2, scoreMin, scoreMax)
		}
		if totalOrders >= toFloat64(frequencyCalc["2_3_threshold"]) {
			frequencyScore = scaleScore(3, scoreMin, scoreMax)
		}
		if totalOrders >= toFloat64(frequencyCalc["3_4_threshold"]) {
			frequencyScore = scaleScore(4, scoreMin, scoreMax)
		}
		if totalOrders >= toFloat64(frequencyCalc["4_5_threshold"]) {
			frequencyScore = scaleScore(5, scoreMin, scoreMax)
		}
	}

	// Calculate Monetary score
	monetaryCalc := extractMap(options["monetary_calculation"])
	monetaryScore := scoreMin
	if percentileBasedScoring {
		monetaryScore = calculatePercentileScore(totalSpendRounded, 0, 10000, scoreMin, scoreMax, false)
	} else {
		// Fixed thresholds
		if totalSpendRounded >= toFloat64(monetaryCalc["1_2_threshold"]) {
			monetaryScore = scaleScore(2, scoreMin, scoreMax)
		}
		if totalSpendRounded >= toFloat64(monetaryCalc["2_3_threshold"]) {
			monetaryScore = scaleScore(3, scoreMin, scoreMax)
		}
		if totalSpendRounded >= toFloat64(monetaryCalc["3_4_threshold"]) {
			monetaryScore = scaleScore(4, scoreMin, scoreMax)
		}
		if totalSpendRounded >= toFloat64(monetaryCalc["4_5_threshold"]) {
			monetaryScore = scaleScore(5, scoreMin, scoreMax)
		}
	}

	// Composite score (concatenated digits, e.g. 543 or 10-9-8)
	compositeScore := recencyScore*100 + frequencyScore*10 + monetaryScore

	// Determine RFM segment label
	segmentLabel := determineRFMSegment(recencyScore, frequencyScore, monetaryScore, scoreMax)

	// Build update data based on save_data field mappings
	saveData := extractMap(options["save_data"])
	updateData := map[string]interface{}{}

	fieldMappings := map[string]interface{}{
		"recency_score":         recencyScore,
		"frequency_score":       frequencyScore,
		"monetary_score":        monetaryScore,
		"rfm_composite_score":   compositeScore,
		"rfm_segment_label":     segmentLabel,
		"total_order_value":     totalSpendRounded,
		"average_order_value":   averageOrderSize,
		"total_order_count":     int(totalOrders),
		"first_order_date":      firstOrderDate,
		"first_order_value":     firstOrderValue,
		"last_order_date":       lastOrderDate,
		"last_order_value":      lastOrderValue,
		"days_since_last_order": daysSinceLastOrder,
	}

	for metricKey, metricValue := range fieldMappings {
		if fieldName, ok := saveData[metricKey].(string); ok && fieldName != "" && fieldName != "no_save" {
			valueStr := fmt.Sprintf("%v", metricValue)
			updateData[fieldName] = valueStr
		}
	}

	// Update contact fields
	for field, value := range updateData {
		err := input.Connector.SetContactFieldValue(ctx, input.ContactID, field, value)
		if err != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to set field '%s': %v", field, err))
		} else {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "field_updated",
				Target: field,
				Value:  value,
			})
		}
	}

	// Apply segment-specific tags if configured
	if segmentLabelsEnabled {
		segmentTags := extractMap(options["segment_tags"])
		segmentTagMap := map[string]string{
			"Champions":       "champions_tag",
			"Loyal":           "loyal_tag",
			"At Risk":         "at_risk_tag",
			"Hibernating":     "hibernating_tag",
			"Lost":            "lost_tag",
			"Promising":       "promising_tag",
			"Need Attention":  "need_attention_tag",
		}

		for segment, tagField := range segmentTagMap {
			if tagID, ok := segmentTags[tagField].(string); ok && tagID != "" {
				if segment == segmentLabel {
					if err := input.Connector.ApplyTag(ctx, input.ContactID, tagID); err != nil {
						output.Logs = append(output.Logs, fmt.Sprintf("Failed to apply %s tag: %v", segment, err))
					} else {
						output.Actions = append(output.Actions, helpers.HelperAction{
							Type:   "tag_applied",
							Target: tagID,
							Value:  segment,
						})
						output.Logs = append(output.Logs, fmt.Sprintf("Applied %s segment tag (%s)", segment, tagID))
					}
				} else {
					if err := input.Connector.RemoveTag(ctx, input.ContactID, tagID); err == nil {
						output.Logs = append(output.Logs, fmt.Sprintf("Removed %s segment tag (%s)", segment, tagID))
					}
				}
			}
		}
	}

	// Fire RFM score goals
	integration := "myfusionhelper"
	helperID := input.HelperID

	goalNames := []string{
		fmt.Sprintf("%srecency%d", helperID, recencyScore),
		fmt.Sprintf("%sfrequency%d", helperID, frequencyScore),
		fmt.Sprintf("%smonetary%d", helperID, monetaryScore),
	}

	for _, goalName := range goalNames {
		goalErr := input.Connector.AchieveGoal(ctx, input.ContactID, goalName, integration)
		if goalErr != nil {
			output.Logs = append(output.Logs, fmt.Sprintf("Failed to achieve goal '%s': %v", goalName, goalErr))
		} else {
			output.Actions = append(output.Actions, helpers.HelperAction{
				Type:   "goal_achieved",
				Target: input.ContactID,
				Value:  goalName,
			})
		}
	}

	output.Success = true
	output.Message = fmt.Sprintf("RFM scores calculated: R=%d F=%d M=%d (composite: %d, segment: %s)", recencyScore, frequencyScore, monetaryScore, compositeScore, segmentLabel)
	output.ModifiedData = map[string]interface{}{
		"recency_score":         recencyScore,
		"frequency_score":       frequencyScore,
		"monetary_score":        monetaryScore,
		"composite_score":       compositeScore,
		"segment_label":         segmentLabel,
		"total_spend":           totalSpendRounded,
		"total_invoice_amount":  totalInvoiceRounded,
		"total_orders":          int(totalOrders),
		"average_order_size":    averageOrderSize,
		"payment_per_order":     actualPaymentPerOrder,
		"days_since_last_order": daysSinceLastOrder,
		"first_order_date":      firstOrderDate,
		"first_order_value":     firstOrderValue,
		"last_order_date":       lastOrderDate,
		"last_order_value":      lastOrderValue,
	}
	output.Logs = append(output.Logs, fmt.Sprintf("RFM for contact %s: R=%d F=%d M=%d composite=%d segment=%s", input.ContactID, recencyScore, frequencyScore, monetaryScore, compositeScore, segmentLabel))

	return output, nil
}

func extractMap(v interface{}) map[string]interface{} {
	if m, ok := v.(map[string]interface{}); ok {
		return m
	}
	return map[string]interface{}{}
}

func toFloat64(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		var f float64
		_, _ = fmt.Sscanf(val, "%f", &f)
		return f
	default:
		var f float64
		_, _ = fmt.Sscanf(fmt.Sprintf("%v", v), "%f", &f)
		return f
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// scaleScore scales a 1-5 score to the configured score range
func scaleScore(score, min, max int) int {
	if max == 5 && min == 1 {
		return score
	}
	// Linear scaling from 1-5 to min-max
	scaled := min + ((score - 1) * (max - min) / 4)
	if scaled < min {
		return min
	}
	if scaled > max {
		return max
	}
	return scaled
}

// calculatePercentileScore calculates a score based on percentile within a range
func calculatePercentileScore(value, rangeMin, rangeMax float64, scoreMin, scoreMax int, inverse bool) int {
	if value <= rangeMin {
		if inverse {
			return scoreMax
		}
		return scoreMin
	}
	if value >= rangeMax {
		if inverse {
			return scoreMin
		}
		return scoreMax
	}

	percentile := (value - rangeMin) / (rangeMax - rangeMin)
	if inverse {
		percentile = 1.0 - percentile
	}

	scoreRange := float64(scoreMax - scoreMin)
	score := scoreMin + int(percentile*scoreRange)

	if score < scoreMin {
		return scoreMin
	}
	if score > scoreMax {
		return scoreMax
	}
	return score
}

// determineRFMSegment assigns a segment label based on RFM scores
func determineRFMSegment(r, f, m, maxScore int) string {
	// Normalize scores to 1-5 scale for segment logic
	rNorm := normalizeToFive(r, maxScore)
	fNorm := normalizeToFive(f, maxScore)
	mNorm := normalizeToFive(m, maxScore)

	// Champions: 5-5-5
	if rNorm >= 5 && fNorm >= 5 && mNorm >= 5 {
		return "Champions"
	}

	// Loyal: R 4-5, F 4-5, M 4-5
	if rNorm >= 4 && fNorm >= 4 && mNorm >= 4 {
		return "Loyal"
	}

	// Promising: R 3-4, F 1-2, M 1-2
	if rNorm >= 3 && rNorm <= 4 && fNorm <= 2 && mNorm <= 2 {
		return "Promising"
	}

	// At Risk: R 2-3, F 2-3, M 4-5
	if rNorm >= 2 && rNorm <= 3 && fNorm >= 2 && fNorm <= 3 && mNorm >= 4 {
		return "At Risk"
	}

	// Hibernating: R 1-2, F 1-2, M 2-3
	if rNorm <= 2 && fNorm <= 2 && mNorm >= 2 && mNorm <= 3 {
		return "Hibernating"
	}

	// Lost: R 1, F 1, M 1-2
	if rNorm <= 1 && fNorm <= 1 && mNorm <= 2 {
		return "Lost"
	}

	// Need Attention: R 2-3, F 2-3, M 2-3
	if rNorm >= 2 && rNorm <= 3 && fNorm >= 2 && fNorm <= 3 && mNorm >= 2 && mNorm <= 3 {
		return "Need Attention"
	}

	// Default
	return "Other"
}

// normalizeToFive normalizes a score from any range to 1-5 scale
func normalizeToFive(score, max int) int {
	if max == 5 {
		return score
	}
	normalized := 1 + ((score - 1) * 4 / (max - 1))
	if normalized < 1 {
		return 1
	}
	if normalized > 5 {
		return 5
	}
	return normalized
}
