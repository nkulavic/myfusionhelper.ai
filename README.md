# MyFusion Helper

AI-powered CRM automation platform. Connect Keap, GoHighLevel, ActiveCampaign, and more. Build powerful automations with AI assistance.

## Architecture

```
myfusionhelper.ai/
├── apps/
│   ├── web/                    # Next.js 15 app (app.myfusionhelper.ai)
│   └── marketing/              # Marketing site (myfusionhelper.ai) [coming soon]
├── backend/
│   └── golang/                 # Go Lambda handlers [coming soon]
├── packages/
│   ├── ui/                     # Shared UI components
│   ├── types/                  # Shared TypeScript types
│   └── config/                 # Shared configs
└── docs/                       # Documentation
```

## Tech Stack

**Frontend:**
- Next.js 15 + React 19
- Tailwind CSS + shadcn/ui
- Better Auth (authentication)
- Zustand (state management)
- Vercel Postgres (database)

**Backend (coming soon):**
- Go + AWS Lambda
- DynamoDB
- SQS queues
- Cognito (for API auth)

## Getting Started

```bash
# Install dependencies
npm install

# Copy environment variables
cp apps/web/.env.example apps/web/.env.local

# Run development server
npm run dev

# Or run just the web app
npm run web
```

## Branches

- `main` - Production
- `staging` - Staging/QA
- `dev` - Development

## Documentation

See [docs/MODERNIZATION_PLAN.md](docs/MODERNIZATION_PLAN.md) for the full roadmap.

## License

Proprietary - MyFusion Solutions
