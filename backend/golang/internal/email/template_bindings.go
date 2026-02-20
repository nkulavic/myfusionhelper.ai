package email

// TemplateDataToBindings converts a TemplateData struct to a flat map
// with snake_case keys for use in Liquid template rendering.
func TemplateDataToBindings(data TemplateData) map[string]interface{} {
	return map[string]interface{}{
		"user_name":       data.UserName,
		"user_email":      data.UserEmail,
		"app_name":        data.AppName,
		"base_url":        data.BaseURL,
		"reset_code":      data.ResetCode,
		"helper_name":     data.HelperName,
		"plan_name":       data.PlanName,
		"error_msg":       data.ErrorMsg,
		"resource_name":   data.ResourceName,
		"usage_percent":   data.UsagePercent,
		"usage_current":   data.UsageCurrent,
		"usage_limit":     data.UsageLimit,
		"total_helpers":   data.TotalHelpers,
		"total_executed":  data.TotalExecuted,
		"total_succeeded": data.TotalSucceeded,
		"total_failed":    data.TotalFailed,
		"success_rate":    data.SuccessRate,
		"top_helper":      data.TopHelper,
		"week_start":      data.WeekStart,
		"week_end":        data.WeekEnd,
		"inviter_name":    data.InviterName,
		"inviter_email":   data.InviterEmail,
		"role_name":       data.RoleName,
		"account_name":    data.AccountName,
		"invite_token":    data.InviteToken,
		"verify_code":     data.VerifyCode,
		"invoice_url":     data.InvoiceURL,
		"card_last4":      data.CardLast4,
		"card_brand":      data.CardBrand,
		"card_exp_month":  data.CardExpMonth,
		"card_exp_year":   data.CardExpYear,
		"amount":          data.Amount,
		"invoice_number":  data.InvoiceNumber,
		"refund_reason":   data.RefundReason,
		"current_year":    getCurrentYear(),
	}
}
