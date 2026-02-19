# MyFusion Helper - Web App (Next.js)

Dashboard app for managing CRM connections, helpers, data exploration, and settings.

## Quick Reference

```bash
npm run web          # dev server on http://localhost:3001
npm run build        # production build
npm run lint         # ESLint
npm run type-check   # tsc --noEmit
```

## Architecture Overview

### App Router (Route Groups)

```
src/app/
├── (auth)/              # Auth pages (no sidebar, centered layout)
│   ├── login/page.tsx
│   ├── register/page.tsx
│   ├── forgot-password/page.tsx
│   └── reset-password/page.tsx
├── (dashboard)/         # Main app (sidebar layout, auth-protected)
│   ├── page.tsx                     # Dashboard home
│   ├── helpers/page.tsx             # Helpers catalog + builder
│   ├── helpers/[id]/page.tsx        # Single helper detail
│   ├── connections/page.tsx         # CRM connections management
│   ├── connections/callback/page.tsx # OAuth callback handler
│   ├── data-explorer/page.tsx       # Query CRM data with filters/NL
│   ├── executions/page.tsx          # Execution history list
│   ├── executions/[id]/page.tsx     # Single execution detail
│   ├── insights/page.tsx            # AI insights
│   ├── reports/page.tsx             # Report list
│   ├── reports/[id]/page.tsx        # Single report
│   ├── emails/page.tsx              # Email management
│   ├── emails/templates/page.tsx    # Email templates
│   ├── plans/page.tsx               # Dedicated plan selection page (Stripe checkout)
│   ├── settings/page.tsx            # Account/profile/API keys/billing
│   └── settings/billing/success/page.tsx # Post-checkout success (confetti)
├── (legal)/             # Legal pages (minimal layout)
│   ├── terms/page.tsx
│   ├── privacy/page.tsx
│   └── eula/page.tsx
├── api/                 # Next.js API routes (server-side proxies)
│   ├── chat/route.ts              # AI chat streaming (Vercel AI SDK)
│   ├── reports/stats/route.ts     # Report stats proxy
│   └── data/                      # Data explorer API proxies
│       ├── catalog/route.ts
│       ├── query/route.ts
│       ├── export/route.ts
│       └── record/[connectionId]/[objectType]/[recordId]/route.ts
├── layout.tsx           # Root layout (Geist font, Providers wrapper)
├── page.tsx             # Landing page (public)
└── globals.css          # CSS variables + Tailwind base
```

### Authentication Flow

1. User submits credentials on `/login`
2. Frontend calls `POST /auth/login` on Go backend
3. Backend authenticates via Cognito `USER_PASSWORD_AUTH`
4. Backend returns `{ token, refresh_token, user, account }`
5. Frontend stores tokens in localStorage (`mfh_access_token`, `mfh_refresh_token`)
6. Frontend sets `mfh_authenticated=1` cookie (for middleware detection)
7. Zustand auth store persists user state (`mfh-auth` in localStorage)

**Middleware** (`src/middleware.ts`):
- Checks for `mfh_authenticated` cookie
- Redirects unauthenticated users to `/login?callbackUrl=...`
- Public routes: `/`, `/login`, `/register`, `/forgot-password`, `/reset-password`, `/terms`, `/privacy`, `/eula`

**Token refresh**: The API client auto-refreshes on 401 responses (single retry with mutex).

### API Client (`src/lib/api/client.ts`)

Central HTTP client with these features:
- Auto-attaches `Authorization: Bearer {token}` header
- Auto-converts request bodies from camelCase to snake_case
- Auto-converts response bodies from snake_case to camelCase
- 401 auto-refresh with single-retry pattern
- Custom `APIError` class with `statusCode`, `code`, `message`

**API modules** (all in `src/lib/api/`):
| Module | Purpose |
|--------|---------|
| `auth.ts` | Login, register, status, logout, refresh |
| `connections.ts` | CRUD connections, platforms, OAuth flow |
| `helpers.ts` | CRUD helpers, types/templates, executions |
| `settings.ts` | Account, API keys, team, notifications, billing |
| `data-explorer.ts` | Catalog, query, record detail, export |
| `emails.ts` | Email CRUD, template CRUD (mock fallback until backend built) |

### State Management

**Zustand stores** (`src/lib/stores/`):

| Store | Key | Persisted | Purpose |
|-------|-----|-----------|---------|
| `auth-store` | `mfh-auth` | user, isAuthenticated | Auth state |
| `workspace-store` | `mfh-workspace` | currentAccount, activeConnectionId | Workspace context |
| `helper-store` | -- | -- | Helper list + builder state |
| `ui-store` | -- | -- | Sidebar, command palette, AI chat |
| `data-explorer-store` | `mfh-data-explorer` | sidebar prefs, page size | Data explorer selection/filters |
| `ai-settings-store` | `mfh-ai-settings` | provider keys, model prefs | AI provider config |

**React Query hooks** (`src/lib/hooks/`):

| File | Key hooks |
|------|-----------|
| `use-auth.ts` | `useAuthStatus` |
| `use-connections.ts` | `useConnections`, `usePlatforms`, `useCreateConnection` |
| `use-helpers.ts` | `useHelpers`, `useCreateHelper`, `useExecutionsPaginated` |
| `use-settings.ts` | `useAccount`, `useUpdateProfile`, `useBillingInfo`, `useInvoices`, `useCreateCheckoutSession`, `useCreatePortalSession`, `useAPIKeys`, `useTeamMembers`, `useNotificationPreferences` |
| `use-emails.ts` | `useEmails`, `useSendEmail`, `useEmailTemplates`, `useCreateTemplate` (mock fallback) |
| `use-plan-limits.ts` | `usePlanLimits` -- returns `isTrialing`, `isTrialExpired`, `daysRemaining`, `canCreate()` with trial-aware logic |
| `use-reports.ts` | `useReportStats` (fetches from `/api/reports/stats`) |

Pattern: each hook wraps an API call with `useQuery` or `useMutation` and handles cache invalidation.

```tsx
// Query hook pattern
export function useHelpers() {
  return useQuery({
    queryKey: ['helpers'],
    queryFn: async () => {
      const res = await helpersApi.list()
      return res.data ?? []
    },
  })
}

// Mutation hook pattern
export function useCreateHelper() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateHelperInput) => helpersApi.create(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['helpers'] })
    },
  })
}
```

**Query defaults** (set in `Providers`):
- `staleTime`: 30 seconds
- `retry`: 1
- `refetchOnWindowFocus`: false

### Component Patterns

**UI components** (`src/components/ui/`): shadcn/ui components. Install new ones with:
```bash
npx shadcn@latest add <component-name>
```

shadcn config (`components.json`):
- Style: default
- RSC: true
- Aliases: `@/components`, `@/components/ui`, `@/lib/utils`

**Feature components** are colocated with their routes using `_components/` directories:
```
src/app/(dashboard)/helpers/
├── page.tsx
└── _components/
    ├── helpers-catalog.tsx
    ├── helper-detail.tsx
    └── helper-builder.tsx
```

**Shared feature components** live in `src/components/`:
- `data-explorer/` -- Data table, filters, JSON viewer, nav tree
- `landing/` -- Landing page sections (hero, features, pricing, etc.)
- `legal/` -- Legal content components

### Form Pattern

Forms use react-hook-form + zod + shadcn Form components:

```tsx
const schema = z.object({
  email: z.string().email('Invalid email'),
  password: z.string().min(1, 'Required'),
})

type FormValues = z.infer<typeof schema>

const form = useForm<FormValues>({
  resolver: zodResolver(schema),
  mode: 'onTouched',
  defaultValues: { email: '', password: '' },
})

// In JSX:
<Form {...form}>
  <form onSubmit={form.handleSubmit(onSubmit)}>
    <FormField
      control={form.control}
      name="email"
      render={({ field }) => (
        <FormItem>
          <FormLabel>Email</FormLabel>
          <FormControl>
            <Input {...field} />
          </FormControl>
          <FormMessage />
        </FormItem>
      )}
    />
  </form>
</Form>
```

### Styling

**Theme system**: CSS variables in HSL format, defined in `globals.css`.

- Light mode: Navy (`212 100% 22%`) primary, light backgrounds
- Dark mode: Lime green (`77 85% 45%`) primary, dark navy backgrounds
- `darkMode: 'class'` controlled by `next-themes`

**Brand tokens**:
- `--brand-navy: 212 100% 22%` (primary brand color)
- `--brand-green: 77 85% 35%` (accent/CTA color)
- `--brand-blue: 220 69% 56%` (info/links)

**Sidebar**: Uses dedicated sidebar CSS variable set (e.g., `--sidebar-background`, `--sidebar-primary`). Navy background in light mode, darker navy in dark mode.

**Utility**: `cn()` function from `src/lib/utils.ts` merges Tailwind classes (clsx + tailwind-merge).

**Fonts**: Geist Sans (body) + Geist Mono (code), loaded via CSS variables.

### AI Chat

- Floating button (bottom-right) opens `AIChatPanel`
- Uses Vercel AI SDK (`ai` package) with streaming
- Provider selection via `ai-settings-store` (Groq free tier default, BYOK for Anthropic/OpenAI)
- Server-side route at `src/app/api/chat/route.ts`

### Dashboard Layout

Fixed 264px sidebar (collapsible to icon-only rail) with:
- Logo (top)
- Account switcher
- Navigation links (9 items + conditional "Plans" link for trial/expired users)
- Theme toggle, notifications, logout, user avatar (bottom)

Main content area: `pl-64` offset, `p-6` padding.

**Trial Banner**: Persistent non-dismissible banner at top of dashboard layout during trial/expired state. Color-coded by urgency: blue (8-14 days), amber (3-7 days), red (1-2 days), dark red (expired). Links to `/plans`.

**Dashboard Tabs**: Dashboard page uses tabbed interface (First Steps / Dashboard):
- **First Steps** tab (default during trial when setup incomplete): trial progress widget, getting started steps (connect CRM, create helper, watch execution), plan CTA
- **Dashboard** tab: stats cards, execution history, quick actions

**Soft Lock**: Expired trial users can only access `/`, `/plans`, `/settings`, `/dashboard`. Other routes redirect to `/plans`.

### Plan Selection Page (`/plans`)

Dedicated page with:
- Monthly/annual billing toggle
- 3 plan cards (Start / Grow / Deliver) with features and pricing
- Feature comparison table (expandable)
- FAQ accordion
- Trust signals ("Cancel anytime", "No charge until trial ends")
- Uses `useCreateCheckoutSession()` for Stripe checkout

### Shared Types

All shared types are in `packages/types/src/index.ts`:
- `User`, `Account`, `UserAccount`, `UserPermissions`
- `NotificationPreferences`, `AccountSettings`, `AccountUsage`
- `PlatformConnection`, `CRMPlatform`
- `Helper`, `HelperExecution`, `HelperTypeDefinition`, `HelperCategory`
- `APIKey`, `NormalizedContact`
- `AuthContext`, `AccountAccess`, `PlanLimits`
- `APIResponse<T>` (standard API envelope)

### Icons

All icons come from `lucide-react`. Do not use other icon libraries.

### Key Libraries

| Library | Version | Purpose |
|---------|---------|---------|
| `next` | ^15.1.6 | Framework |
| `react` | ^19.0.0 | UI |
| `@tanstack/react-query` | ^5.62 | Server state |
| `@tanstack/react-table` | ^8.21 | Data tables |
| `zustand` | ^5.0 | Client state |
| `react-hook-form` | ^7.71 | Forms |
| `zod` | ^4.3 | Validation |
| `framer-motion` | ^12.31 | Animations |
| `ai` / `@ai-sdk/*` | ^6.0 | AI chat |
| `cmdk` | ^1.1 | Command palette |
| `lucide-react` | ^0.468 | Icons |
| `next-themes` | ^0.4 | Dark/light theme |
| `tailwindcss-animate` | ^1.0 | CSS animation utilities |

### Stubs and TODOs

- **Emails**: `use-emails.ts` hooks fall back to mock data when backend email endpoints return errors (backend email CRUD not yet implemented).
- **Settings**: Team members and notification preferences still use stub API calls until backend endpoints are built.
- **Billing**: Wired to real Stripe backend endpoints (`/billing`, `/billing/checkout/sessions`, `/billing/portal-session`, `/billing/invoices`). Billing response includes trial state: `isTrialing`, `daysRemaining`, `totalTrialDays`, `trialExpired`.

### Plan Constants (`src/lib/plan-constants.ts`)

Single source of truth for plan configuration. `PlanId` type: `'free' | 'trial' | 'start' | 'grow' | 'deliver'`. Key helpers: `isTrialPlan()`, `isPaidPlan()`, `getPlanLabel()`, `getPlanFeatures()`, `COMPARISON_ROWS` for feature comparison table.
