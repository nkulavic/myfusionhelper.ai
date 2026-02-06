# MyFusion Helper - End-to-End Test Plan

Manual QA checklist for testing all major user flows before releases.

**App URL**: `https://app.myfusionhelper.ai` (production) / `http://localhost:3001` (local dev)
**API URL**: `https://a95gb181u4.execute-api.us-west-2.amazonaws.com` (dev)

---

## 1. Authentication Flows

### 1.1 Registration

**Preconditions**: No existing account for the test email.

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Navigate to `/register` | Registration form loads with name, email, password fields |
| 2 | Submit with empty fields | Validation errors shown for required fields |
| 3 | Enter invalid email (no @) | "Invalid email format" error |
| 4 | Enter password < 8 chars | "Password must be at least 8 characters" error |
| 5 | Enter phone without `+` prefix | "Phone number must start with +" error |
| 6 | Fill valid fields and submit | Spinner shown, then redirect to `/onboarding` |
| 7 | Check localStorage | `mfh_access_token` and `mfh_refresh_token` set |
| 8 | Check cookie | `mfh_authenticated=1` cookie present |
| 9 | Check Zustand store | `mfh-auth` in localStorage has user object and `isAuthenticated: true` |

**Error scenarios**:
- Register with an already-registered email: should return "Email already registered" (409)
- Network failure during registration: should show generic error message

### 1.2 Login

**Preconditions**: User account exists.

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Navigate to `/login` | Login form with email and password fields |
| 2 | Submit with empty fields | Validation errors for required fields |
| 3 | Enter wrong password | "Invalid credentials" error displayed |
| 4 | Enter correct credentials | Spinner, then redirect to dashboard `/` or `callbackUrl` |
| 5 | Verify tokens stored | `mfh_access_token`, `mfh_refresh_token` in localStorage |
| 6 | Verify sidebar shows user avatar | User name and avatar visible in sidebar |

**Error scenarios**:
- Unconfirmed Cognito user: should show appropriate error
- Account disabled: should show "Account disabled" message
- Network timeout: should show error, not hang

### 1.3 Logout

**Preconditions**: User is logged in.

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Click logout button in sidebar | Redirect to `/login` |
| 2 | Check localStorage | Tokens cleared, `mfh-auth` reset |
| 3 | Check cookie | `mfh_authenticated` cookie removed |
| 4 | Navigate to `/helpers` directly | Redirected to `/login?callbackUrl=/helpers` |

### 1.4 Forgot Password

**Preconditions**: User account exists with verified email.

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Navigate to `/forgot-password` | Form with email field shown |
| 2 | Submit with empty email | Validation error |
| 3 | Submit with valid email | Success message: "Check your email for a verification code" |
| 4 | Submit with non-existent email | Same success message (no information leakage) |
| 5 | Click "Back to login" link | Navigate to `/login` |

### 1.5 Reset Password

**Preconditions**: Verification code received via email from forgot-password flow.

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Navigate to `/reset-password` | Form with email, code, new password, confirm password |
| 2 | Enter password without uppercase | Validation: "must contain an uppercase letter" |
| 3 | Enter password without number | Validation: "must contain a number" |
| 4 | Enter mismatched passwords | Validation: "Passwords do not match" |
| 5 | Submit with valid fields | Success message, link to login |
| 6 | Log in with new password | Login succeeds |

**Error scenarios**:
- Expired verification code: should show "Code expired" error
- Invalid code: should show "Invalid code" error

### 1.6 Token Refresh

**Preconditions**: User is logged in, access token is expired.

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Wait for token to expire (or manually clear access token) | Next API call triggers 401 |
| 2 | API client auto-refreshes | New access token obtained via refresh token |
| 3 | Original request retried | Request succeeds with new token |
| 4 | If refresh token also expired | User redirected to `/login` |

### 1.7 Auth Middleware

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Access `/helpers` without auth cookie | Redirect to `/login?callbackUrl=/helpers` |
| 2 | Access `/login` while authenticated | Page loads (not redirected) |
| 3 | Access `/` (landing) without auth | Page loads normally |
| 4 | Access `/terms`, `/privacy`, `/eula` without auth | Pages load normally |
| 5 | Access `/onboarding` without auth | Redirect to login |

---

## 2. Onboarding Flow

**Preconditions**: User just registered (or `onboardingComplete` is false in workspace store).

### 2.1 Welcome Step

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | User redirected to `/onboarding` after registration | Welcome step shown with user's first name |
| 2 | Progress indicator shows step 1 of 4 | Step 1 circle highlighted |
| 3 | Click "Get Started" | Animate to Connect CRM step |
| 4 | Click "Skip setup" | Redirect to `/helpers`, onboarding marked complete |

### 2.2 Connect CRM Step

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Platform list displayed | Shows Keap, GoHighLevel, ActiveCampaign, Ontraport, HubSpot |
| 2 | Click an OAuth platform (e.g., Keap) | OAuth flow initiated, redirect to CRM auth page |
| 3 | Click an API Key platform (e.g., ActiveCampaign) | API key input form shown |
| 4 | Click "Skip for now" | Advance to Pick Helper step |
| 5 | Click "Back" | Return to Welcome step |

### 2.3 Pick Helper Step

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Popular helpers shown in grid | 8 popular helpers displayed |
| 2 | Click category tabs | Filtered helpers shown for that category |
| 3 | Click a helper card | Card highlighted with checkmark, count updated |
| 4 | Click highlighted card again | Deselected |
| 5 | Click "Continue" with selections | Loading spinner, helpers created via API, advance to tour |
| 6 | Click "Skip & Continue" with no selections | Advance without creating helpers |

**Error scenarios**:
- CRM-dependent helpers selected with no connection: shows warning message
- API failure creating helper: continues with remaining helpers

### 2.4 Quick Tour Step

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Tour highlights shown | Key features overview displayed |
| 2 | Click "Let's Go" / finish button | Redirect to `/helpers`, onboarding marked complete |
| 3 | Click "Back" | Return to Pick Helper step |

### 2.5 Already Completed

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Navigate to `/onboarding` after completing | Redirect to `/helpers` |

---

## 3. Connections

**Preconditions**: User is logged in with an active account.

### 3.1 View Connections

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Navigate to `/connections` | List of existing connections shown (or empty state) |
| 2 | Empty state | "No connections yet" message with "Add Connection" button |
| 3 | Connection cards show status | Green "Connected" badge or red "Error" badge |

### 3.2 Add OAuth Connection (e.g., Keap)

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Click "Add Connection" | Platform selection view shown |
| 2 | Select Keap | OAuth flow description shown |
| 3 | Click "Connect" | Redirect to Keap OAuth authorization page |
| 4 | Authorize in Keap | Redirect back to `/connections/callback?code=...&state=...` |
| 5 | Callback processed | Connection created, redirect to connections list |
| 6 | New connection appears | Keap connection shown with "Connected" status |

**Error scenarios**:
- User denies OAuth access: callback page shows error, back to connections
- Invalid state parameter: error shown, no connection created

### 3.3 Add API Key Connection (e.g., ActiveCampaign)

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Select ActiveCampaign | API key input form shown with URL and key fields |
| 2 | Enter invalid API key | "Connection failed" error after test |
| 3 | Enter valid API key and URL | Connection created, appears in list |

### 3.4 Test Connection

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Click connection card to view detail | Connection detail view shown |
| 2 | Click "Test Connection" | Spinner, then success or failure message |

### 3.5 Delete Connection

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Click delete button on connection | Confirmation dialog shown |
| 2 | Confirm delete | Connection removed from list |
| 3 | Cancel delete | Dialog closes, connection remains |

---

## 4. Helpers

**Preconditions**: User is logged in, at least one CRM connection exists.

### 4.1 Browse Catalog

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Navigate to `/helpers` | "My Helpers" tab shown (default) |
| 2 | Switch to "Catalog" tab | Full catalog of available helper types |
| 3 | Filter by category (contact, data, tagging, etc.) | Filtered results shown |
| 4 | Search helpers | Results filter by name match |

### 4.2 Create Helper

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Click a helper type in catalog | Helper builder opens with `?view=new&type=tag_it` |
| 2 | Fill in helper name | Name field populated |
| 3 | Select CRM connection | Connection dropdown shows available connections |
| 4 | Configure helper (if config form available) | Type-specific form shown (e.g., tag selection for tag_it) |
| 5 | Click "Create Helper" | Spinner, helper created, redirect to my helpers list |
| 6 | New helper appears in "My Helpers" | Helper card shown with "Active" status |

**Error scenarios**:
- Missing required fields: validation prevents submission
- API failure: error message shown

### 4.3 Configure Helper (Config Forms)

Test each available config form:

| Helper Type | Config Form | Key Fields to Test |
|------------|-------------|-------------------|
| tag_it | TagItForm | Action (apply/remove/toggle), tag chips, condition |
| copy_it | CopyItForm | Source field, target field, overwrite toggle |
| format_it | FormatItForm | Format type dropdown, field chips |
| hook_it | HookItForm | Webhook URL, HTTP method, custom headers, include contact toggle |
| clear_tags | ClearTagsForm | Mode (all/prefix/category), conditional fields, warning for "all" |
| math_it | MathItForm | Source field, operation, value, target field |
| notify_me | NotifyMeForm | Channel (email/slack/webhook), conditional fields per channel |

For each form:
- Verify disabled state when `disabled` prop is true
- Verify labels have proper `htmlFor` attributes
- Verify Enter key adds items (for chip inputs)
- Verify items can be removed

### 4.4 Activate / Deactivate Helper

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Toggle helper status | Status changes to active/inactive |
| 2 | Inactive helper | Should not execute |

### 4.5 Delete Helper

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Click delete on helper | Confirmation dialog |
| 2 | Confirm | Helper removed from list |

---

## 5. Executions

**Preconditions**: User has executed at least one helper.

### 5.1 View Execution List

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Navigate to `/executions` | List of executions shown with status icons |
| 2 | Each row shows | Helper name, status badge, duration, timestamp |
| 3 | Empty state | "No executions yet" message |

### 5.2 Filter by Status

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Click "Completed" filter | Only completed executions shown |
| 2 | Click "Failed" filter | Only failed executions shown |
| 3 | Click "Running" filter | Only running executions shown |
| 4 | Click "All" filter | All executions shown |

### 5.3 Pagination

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | If > 20 executions, "Next" button appears | Clicking loads next page |
| 2 | On page 2+, "Previous" button appears | Clicking returns to previous page |
| 3 | Page indicator shows current position | Correct page number displayed |

### 5.4 Execution Detail

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Click an execution row | Navigate to `/executions/[id]` |
| 2 | Detail page shows | Helper name, status, start/end time, duration, contact info |
| 3 | For failed executions | Error message and stack trace shown |
| 4 | Back button | Returns to execution list |

---

## 6. Data Explorer

**Preconditions**: User has at least one active CRM connection with synced data.

### 6.1 Select Connection and Object

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Navigate to `/data-explorer` | Sidebar with hierarchical connection/object nav |
| 2 | Expand a connection | Object types shown (Contacts, Tags, etc.) |
| 3 | Click an object type | Data table loads in main content area |

### 6.2 Run Query

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Select contacts object | Contact records displayed in table |
| 2 | Click a row | Record detail view or JSON viewer shown |

### 6.3 Apply Filters

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Add a filter (e.g., email contains "gmail") | Table updates with filtered results |
| 2 | Add multiple filters | Results narrow with each filter |
| 3 | Clear filters | Full dataset restored |

### 6.4 Export Data

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Click "Export" button | Export dialog or download initiated |
| 2 | Select format (CSV/JSON) | File downloaded in selected format |

### 6.5 Natural Language Query

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Enter NL query (e.g., "contacts tagged as VIP") | AI translates to structured query |
| 2 | Results shown | Matching records displayed |

### 6.6 Sidebar Resize

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Drag sidebar divider | Sidebar width changes (min 200px, max 600px) |
| 2 | Toggle sidebar closed | Sidebar collapses, main content expands |
| 3 | Refresh page | Sidebar width persisted via Zustand store |

---

## 7. Reports

**Preconditions**: User has execution history data.

### 7.1 View Overview

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Navigate to `/reports` | Overview tab active by default |
| 2 | KPI cards shown | Total Executions, Success Rate, Avg Duration, Failed count |
| 3 | Daily trend chart | Bar chart showing last 30 days |
| 4 | Hover on bar | Tooltip with date, total, failed counts |
| 5 | Top helpers list | Ranked by execution count with success rate |
| 6 | Error breakdown | Recent errors with count badges (only if errors exist) |

### 7.2 Report Types Tab

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Switch to "Report Types" tab | 6 report template cards shown |
| 2 | Filter by category | Cards filter by overview/performance/contacts |
| 3 | Search reports | Cards filter by name match |
| 4 | Click a report card | Navigate to `/reports/[id]` |

### 7.3 Report Detail

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | View execution-overview report | KPI grid (6 metrics), stacked bar chart, helper table, error breakdown |
| 2 | View helper-performance report | KPI grid, helper performance table with columns |
| 3 | View error-analysis report | KPI grid, error breakdown with percentage bars |
| 4 | View contact-activity report | Contact processing summary, performance stats |
| 5 | Back arrow | Returns to `/reports` |

**Error scenarios**:
- No execution data: empty state with helpful message
- API unavailable: error state shown

---

## 8. Emails

**Preconditions**: User is logged in.

### 8.1 Compose Email

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Navigate to `/emails` | Compose tab active by default |
| 2 | Enter recipient in "To" field | Text input accepted |
| 3 | Enter subject line | Text input accepted |
| 4 | Write email body | Textarea accepts text, supports {{tokens}} |
| 5 | Formatting toolbar | Bold, italic, link, list, align buttons shown |
| 6 | Click personalization token | Token appended to body |
| 7 | Select tone in AI Composer | Tone pill highlighted |

### 8.2 Send Email

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Fill to, subject, body | Send button enabled |
| 2 | Click "Send Email" | Loading spinner, then success |
| 3 | After send | Fields cleared, switched to "Sent & Drafts" tab |
| 4 | Send with empty subject | Button disabled (cannot submit) |
| 5 | Send with empty "to" | Button disabled (cannot submit) |

### 8.3 Sent & Drafts Tab

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Switch to "Sent & Drafts" tab | Email list shown with status badges |
| 2 | Filter by status (sent/scheduled/draft/failed) | List filters correctly |
| 3 | Search emails | Filter by subject or recipient |
| 4 | Sent emails show metrics | Open count and click count displayed |

### 8.4 Delete Email

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Hover over email row | Delete icon appears |
| 2 | Click delete icon | Confirmation dialog shown |
| 3 | Confirm delete | Email removed from list |
| 4 | Cancel | Dialog closes, email remains |

### 8.5 Email Templates

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Navigate to `/emails/templates` | Template grid shown |
| 2 | Search templates | Filter by name or subject |
| 3 | Filter by category | Buttons filter by onboarding/sales/billing/marketing/events |

### 8.6 Create Template

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Click "New Template" | Create dialog opens |
| 2 | Fill name, category, subject, body | All fields accept input |
| 3 | Click "Create Template" | Spinner, dialog closes, template appears in grid |
| 4 | Submit with empty name | Button disabled |
| 5 | Submit with empty subject | Button disabled |

### 8.7 Edit Template

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Hover over template card | Edit icon appears |
| 2 | Click edit icon | Edit dialog opens with pre-filled fields |
| 3 | Modify fields and save | Changes persisted, dialog closes |

### 8.8 Delete Template

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Hover over template card | Delete icon appears |
| 2 | Click delete icon | Confirmation dialog with template name |
| 3 | Confirm delete | Template removed from grid |

### 8.9 Star Template

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Click star icon on template | Star fills with color |
| 2 | Click again | Star unfills (toggled) |

---

## 9. Settings

**Preconditions**: User is logged in.

### 9.1 Profile Tab

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Navigate to `/settings` | Profile tab shown by default |
| 2 | Name and email pre-filled | From auth store user data |
| 3 | Modify name | Save button enabled |
| 4 | Click "Save Changes" | Spinner, profile updated via PUT /auth/profile |
| 5 | Verify auth store updated | `updateUserData` called with new name/email |

### 9.2 Account Tab

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Switch to Account tab | Company name, timezone fields shown |
| 2 | Update company name | Save persists via PUT /accounts/{id} |

### 9.3 API Keys Tab

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Switch to API Keys tab | List of existing keys (or empty state) |
| 2 | Click "Create API Key" | Dialog with name and permissions |
| 3 | Create key | Key shown once (copy opportunity), then masked in list |
| 4 | Copy key | Copied to clipboard |
| 5 | Click "Revoke" on a key | Confirmation, then key removed |

### 9.4 Billing Tab

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Switch to Billing tab | Current plan card shown |
| 2 | Plan display | Plan name, price, renewal date, trial info if applicable |
| 3 | Usage section | Progress bars for executions, helpers, connections, API keys |
| 4 | Invoice history | List of past invoices with status badges |

### 9.5 Upgrade Plan (Stripe Checkout)

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Click plan tier card (e.g., "Grow") | Loading spinner on that card |
| 2 | POST /billing/checkout/sessions called | Returns `{ url, sessionId }` |
| 3 | Browser redirects to Stripe Checkout URL | Stripe checkout page loads (same tab) |
| 4 | Complete payment in Stripe | Redirect to `/settings?tab=billing&session_id=...` |
| 5 | Or navigate to `/settings/billing/success` | Confetti animation, success message |

**Error scenarios**:
- Checkout session creation fails: error message shown below plan cards
- Invalid plan ID: 400 error from backend

### 9.6 Manage Subscription (Stripe Portal)

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Click "Manage Subscription" (only shown for paid plans) | Loading spinner |
| 2 | POST /billing/portal-session called | Returns `{ url }` |
| 3 | Browser redirects to Stripe Portal | Portal page loads (same tab) |
| 4 | After portal actions | Redirect back to `/settings?tab=billing` |

### 9.7 Team Tab

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Switch to Team tab | Team member list shown |
| 2 | Current user shown | With "Owner" role badge |
| 3 | Click "Invite Member" | Invite form expands with email and role fields |
| 4 | Enter email and select role | Submit sends invite via API |
| 5 | New member appears | With "Pending" status |

### 9.8 Notifications Tab

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Switch to Notifications tab | Toggle switches for notification types |
| 2 | Toggle a preference | Switch animates, preference saved via API |

### 9.9 AI Settings Tab

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Switch to AI tab | Provider selection and API key input |
| 2 | Select provider (Groq/Anthropic/OpenAI) | Model dropdown updates per provider |
| 3 | Enter API key | Key persisted in Zustand store (localStorage) |

---

## 10. Billing End-to-End Flow

**Preconditions**: User on "Free" plan, Stripe test mode configured.

### 10.1 New Subscription

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Go to Settings > Billing | Free plan shown, no "Manage Subscription" button |
| 2 | Click "Get Started" on Start plan ($39) | Checkout session created |
| 3 | Redirected to Stripe Checkout | Pre-filled with customer email, Start plan line item |
| 4 | Enter test card `4242 4242 4242 4242` | Payment succeeds |
| 5 | Redirected back to app | Success page or billing tab |
| 6 | Verify plan updated | Billing tab shows "Start Plan" with renewal date |
| 7 | Verify DynamoDB | Account record updated with plan=start, limits updated |
| 8 | Check email | Welcome/activation billing email received |

### 10.2 Plan Upgrade

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | On Start plan, click "Upgrade" on Grow plan | Checkout session created |
| 2 | Complete Stripe Checkout | New subscription created |
| 3 | Webhook: `checkout.session.completed` fires | Account updated to Grow plan |
| 4 | Webhook: `customer.subscription.created` fires | Plan limits updated |
| 5 | Billing tab reflects new plan | "Grow Plan", $59/month shown |

### 10.3 Manage via Stripe Portal

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Click "Manage Subscription" | Redirect to Stripe Customer Portal |
| 2 | Cancel subscription in portal | Webhook: `customer.subscription.deleted` fires |
| 3 | Return to app | Plan downgraded to Free |
| 4 | Verify limits reset | MaxHelpers=3, MaxConnections=1, MaxExecutions=1000 |

### 10.4 Payment Failure

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Stripe sends `invoice.payment_failed` webhook | Logged, no immediate plan change |
| 2 | Stripe retries payment | `customer.subscription.updated` with `past_due` status |
| 3 | Account remains active | Status stays "active" (past_due treated as active) |
| 4 | If all retries fail | `customer.subscription.deleted` fires, plan reset to Free |

### 10.5 Webhook Signature Verification

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | Send webhook without `stripe-signature` header | 400 "Missing Stripe signature" |
| 2 | Send webhook with invalid signature | 400 "Invalid signature" |
| 3 | Send webhook with valid signature | 200 OK, event processed |

### 10.6 Trial Period

| # | Step | Expected Result |
|---|------|-----------------|
| 1 | New checkout includes 14-day trial | Subscription created with `trialing` status |
| 2 | During trial | Account has full plan access |
| 3 | Trial ends, payment succeeds | Subscription moves to `active` |
| 4 | Trial ends, payment fails | Subscription moves to `past_due` |

---

## Cross-Cutting Concerns

### Responsive Design

| # | Check | Expected |
|---|-------|----------|
| 1 | All pages at 1440px width | Full layout, sidebar visible |
| 2 | All pages at 1024px width | Layout adapts, no overflow |
| 3 | All pages at 768px width | Sidebar collapsible, content stacks |
| 4 | All pages at 375px width | Mobile-friendly, no horizontal scroll |

### Dark Mode

| # | Check | Expected |
|---|-------|----------|
| 1 | Toggle theme in sidebar | All pages switch to dark mode |
| 2 | Check all cards, inputs, buttons | Proper contrast, readable text |
| 3 | Charts and visualizations | Visible in both modes |
| 4 | Status badges | Colors distinguishable in both modes |

### Loading States

| # | Check | Expected |
|---|-------|----------|
| 1 | Every data-fetching page | Shows skeleton or spinner while loading |
| 2 | Mutation buttons (save, create, delete) | Show spinner during pending state |
| 3 | Buttons disabled during mutations | Prevent double-submission |

### Error States

| # | Check | Expected |
|---|-------|----------|
| 1 | API unavailable | Graceful error message, not blank page |
| 2 | 401 during session | Auto-refresh attempted, then redirect to login |
| 3 | Network offline | Appropriate error handling |

### Empty States

| # | Check | Expected |
|---|-------|----------|
| 1 | No helpers | "No helpers yet" with CTA to create |
| 2 | No connections | "No connections" with CTA to add |
| 3 | No executions | "No executions yet" message |
| 4 | No emails | "Compose and send your first email" |
| 5 | No report data | "Execute a helper to see analytics" |

### Accessibility

| # | Check | Expected |
|---|-------|----------|
| 1 | All icon-only buttons | Have `aria-label` attributes |
| 2 | Form inputs | Have associated `<Label>` with `htmlFor` |
| 3 | Dialogs and alerts | Proper focus management |
| 4 | Keyboard navigation | Tab order logical, Enter/Space activate buttons |
| 5 | Color contrast | Meets WCAG AA for text on backgrounds |

---

## Environment Setup for Testing

### Test Stripe Cards

| Card Number | Scenario |
|------------|----------|
| `4242 4242 4242 4242` | Successful payment |
| `4000 0000 0000 0002` | Card declined |
| `4000 0000 0000 3220` | 3D Secure required |

Use any future expiration date, any 3-digit CVC, any postal code.

### Test Accounts

Create test accounts for each plan tier:
- Free plan user (default after registration)
- Start plan user (after checkout)
- Grow plan user (after checkout)
- Deliver plan user (after checkout)

### DynamoDB Verification

After billing changes, verify the accounts table directly:
```bash
aws dynamodb get-item \
  --table-name mfh-dev-accounts \
  --key '{"account_id": {"S": "account:YOUR_ID"}}' \
  --projection-expression "#p, #s, settings" \
  --expression-attribute-names '{"#p": "plan", "#s": "status"}' \
  --region us-west-2
```

### Webhook Testing

Use Stripe CLI for local webhook testing:
```bash
stripe listen --forward-to localhost:3000/billing/webhook
stripe trigger checkout.session.completed
stripe trigger customer.subscription.updated
stripe trigger customer.subscription.deleted
stripe trigger invoice.payment_failed
```
