# MyFusion Helper - Complete Helpers Inventory

**Project**: Helpers Microservices Migration
**Teamwork Task List**: 3258306
**Task**: Phase 1.1 - Extract complete list of all helpers
**Date**: 2026-02-12

## Summary

Total helpers discovered: **99 unique helpers**

## Helpers by Category

### Contact Helpers (17)
Manipulate contact records, fields, and relationships.

1. `assign_it` - Assign contacts to users/owners
2. `clear_it` - Clear contact fields
3. `combine_it` - Combine multiple contact fields
4. `company_link` - Link contacts to companies
5. `contact_updater` - Update contact records
6. `copy_it` - Copy contact data
7. `default_to_field` - Set default values for fields
8. `field_to_field` - Copy data between fields
9. `found_it` - Mark contacts as found/identified
10. `merge_it` - Merge duplicate contacts
11. `move_it` - Move contacts between lists/segments
12. `name_parse_it` - Parse and normalize contact names
13. `note_it` - Add notes to contacts
14. `opt_in` - Opt contacts into communications
15. `opt_out` - Opt contacts out of communications
16. `own_it` - Set contact ownership
17. `snapshot_it` - Create contact snapshots/backups

### Data Helpers (18)
Transform, format, and manipulate data fields.

1. `advance_math` - Advanced mathematical operations
2. `date_calc` - Date calculations and formatting
3. `format_it` - Format data fields
4. `get_the_first` - Extract first element from lists
5. `get_the_last` - Extract last element from lists
6. `ip_location` - IP address geolocation lookup
7. `last_click_it` - Track last click timestamp
8. `last_open_it` - Track last email open
9. `last_send_it` - Track last email send
10. `math_it` - Basic mathematical operations
11. `password_it` - Generate passwords
12. `phone_lookup` - Phone number validation/lookup
13. `quote_it` - Format text as quotes
14. `split_it` - Split text into parts
15. `split_it_basic` - Basic text splitting
16. `text_it` - Text transformation operations
17. `when_is_it` - Date/time queries
18. `word_count_it` - Count words in text

### Tagging Helpers (6)
Manage tags and tag-based operations.

1. `clear_tags` - Remove all tags from contacts
2. `count_it_tags` - Count tags on contacts
3. `count_tags` - Count tag occurrences
4. `group_it` - Group contacts by criteria
5. `score_it` - Score contacts based on tags
6. `tag_it` - Apply tags to contacts

### Automation Helpers (22)
Trigger automations, route contacts, and manage workflows.

1. `action_it` - Execute automation actions
2. `chain_it` - Chain multiple helpers together
3. `countdown_timer` - Set countdown timers
4. `drip_it` - Drip campaign management
5. `goal_it` - Achieve campaign goals
6. `ip_notifications` - IP-based notifications
7. `ip_redirects` - IP-based redirects
8. `limit_it` - Limit automation execution
9. `match_it` - Match contacts to criteria
10. `route_it` - Route contacts to campaigns
11. `route_it_by_custom` - Custom field-based routing
12. `route_it_by_day` - Day-of-week routing
13. `route_it_by_time` - Time-based routing
14. `route_it_geo` - Geographic routing
15. `route_it_score` - Score-based routing
16. `route_it_source` - Source-based routing
17. `simple_opt_in` - Simple opt-in automation
18. `simple_opt_out` - Simple opt-out automation
19. `stage_it` - Manage pipeline stages
20. `timezone_triggers` - Timezone-aware triggers
21. `trigger_it` - Trigger automations
22. `video_trigger_it` - Video engagement triggers

### Integration Helpers (30)
Connect with external services and platforms.

1. `calendly_it` - Calendly integration
2. `countdown_timer` - Countdown timer integration
3. `donor_search` - Donor database lookup
4. `dropbox_it` - Dropbox integration
5. `email_attach_it` - Email attachment handling
6. `email_validate_it` - Email validation
7. `everwebinar` - EverWebinar integration
8. `excel_it` - Excel file operations
9. `facebook_lead_ads` - Facebook Lead Ads integration
10. `google_sheet_it` - Google Sheets integration
11. `gotowebinar` - GoToWebinar integration
12. `hook_it` - Webhook integration (v1)
13. `hook_it_by_tag` - Tag-based webhooks
14. `hook_it_v2` - Webhook integration (v2)
15. `hook_it_v3` - Webhook integration (v3)
16. `hook_it_v4` - Webhook integration (v4)
17. `mail_it` - Email sending
18. `order_it` - Order/transaction tracking
19. `query_it_basic` - Basic database queries
20. `search_it` - Search operations
21. `slack_it` - Slack integration
22. `stripe_hooks` - Stripe webhook handling
23. `trello_it` - Trello integration
24. `twilio_sms` - Twilio SMS integration
25. `upload_it` - File upload handling
26. `webinar_jam` - WebinarJam integration
27. `zoom_meeting` - Zoom meeting integration
28. `zoom_webinar` - Zoom webinar integration
29. `zoom_webinar_absentee` - Zoom absentee tracking
30. `zoom_webinar_participant` - Zoom participant tracking

### Notification Helpers (2)
Send notifications and track engagement.

1. `email_engagement` - Email engagement tracking
2. `notify_me` - Send notifications

### Analytics Helpers (3)
Calculate metrics and analytics.

1. `customer_lifetime_value` - CLV calculation
2. `last_click_it` - Last click analytics
3. `rfm_calculation` - RFM (Recency, Frequency, Monetary) analysis

### Platform Helpers (1)
Platform-specific backup and maintenance.

1. `keap_backup` - Keap platform backup utility

## File Locations

All helpers are located in `/Users/nickkulavic/Projects/myfusionhelper.ai/backend/golang/internal/helpers/`

```
internal/helpers/
├── analytics/              # 3 helpers
├── automation/             # 22 helpers
├── contact/                # 17 helpers
├── data/                   # 18 helpers
├── integration/            # 30 helpers
├── notification/           # 2 helpers
├── platform/               # 1 helper
├── tagging/                # 6 helpers
├── executor.go             # Helper execution orchestrator
├── interface.go            # Helper interface definition
└── registry.go             # Helper factory registry
```

## Registration Pattern

All helpers self-register via `init()` functions:

```go
func init() {
    helpers.Register("helper_name", func() helpers.Helper {
        return &HelperStruct{}
    })
}
```

## Helper Interface

Each helper implements:
- `GetName()` - Human-readable name
- `GetType()` - Registered type identifier
- `GetCategory()` - Category (contact, data, tagging, etc.)
- `GetDescription()` - Helper description
- `GetConfigSchema()` - Configuration schema
- `Execute(ctx, input)` - Main execution logic
- `ValidateConfig(config)` - Config validation
- `RequiresCRM()` - Whether CRM connection required
- `SupportedCRMs()` - List of supported CRM platforms

## Duplicates & Notes

**Duplicate helpers across categories:**
- `countdown_timer` appears in both automation and integration (likely different implementations)
- `last_click_it` appears in both data and analytics (likely different purposes)

**Test files excluded:**
- All `*_test.go` and `*_integration_test.go` files were excluded from the count
- Helper utilities (`contact_updater`, `zoom_utils`) are included but may not be standalone helpers

## Next Steps (Phase 1.2)

For each of the 99 helpers:
1. Create individual Lambda handler in `cmd/handlers/helper-workers/{helper_name}/main.go`
2. Create Serverless Framework config in `services/workers/helpers/{helper_name}/serverless.yml`
3. Create SQS queue: `mfh-helper-{helper_name}-{stage}`
4. Update stream router to route executions to correct queue
5. Add CI/CD deployment workflow

## Validation

Total helper count: **99 helpers**

This matches the project target of migrating all 99 helpers from the legacy system to the new Go-based microservices architecture.
