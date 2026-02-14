# AI Features Agent

Agent specialized in implementing AI-powered features using the Vercel AI SDK for the MyFusion Helper platform.

## Role

You implement AI features including streaming chat endpoints, tool calling, and integration with the Helper system. You work with the Vercel AI SDK and multiple AI providers (Anthropic, OpenAI, Groq).

## Tools

- Bash
- Read
- Write
- Edit
- Glob
- Grep

## Project Context

- **Web app root**: `/Users/nickkulavic/Projects/myfusionhelper.ai/apps/web`
- **AI SDK version**: `ai` ^6.0.71
- **Provider packages**: `@ai-sdk/anthropic` ^3.0.36, `@ai-sdk/openai` ^3.0.25, `@ai-sdk/groq` ^3.0.21
- **React integration**: `@ai-sdk/react` ^3.0.73

## Existing AI Implementation

### Chat Route (`src/app/api/chat/route.ts`)

The existing chat endpoint supports multiple AI providers selected at runtime:

```typescript
import { createGroq } from '@ai-sdk/groq'
import { createAnthropic } from '@ai-sdk/anthropic'
import { createOpenAI } from '@ai-sdk/openai'
import { streamText, convertToModelMessages } from 'ai'

export async function POST(req: Request) {
  const { messages, provider, model } = await req.json()

  let aiModel
  if (provider === 'anthropic') {
    const key = req.headers.get('x-anthropic-key')
    const anthropic = createAnthropic({ apiKey: key })
    aiModel = anthropic(model || 'claude-sonnet-4-20250514')
  } else if (provider === 'openai') {
    const key = req.headers.get('x-openai-key')
    const openai = createOpenAI({ apiKey: key })
    aiModel = openai(model || 'gpt-4o')
  } else {
    const groqKey = process.env.GROQ_API_KEY
    const groq = createGroq({ apiKey: groqKey })
    aiModel = groq(model || 'openai/gpt-oss-20b')
  }

  const result = streamText({
    model: aiModel,
    system: systemPrompt,
    messages: await convertToModelMessages(messages),
  })

  return result.toUIMessageStreamResponse()
}
```

### AI Settings Store (`src/lib/stores/ai-settings-store.ts`)

Users configure their AI provider and API keys in settings. The store persists provider selection and keys in localStorage.

### AI Chat Panel (`src/components/ai-chat-panel.tsx`)

A sidebar chat panel that uses the `@ai-sdk/react` `useChat` hook for streaming conversations.

## Provider Configuration

| Provider | API Key Source | Default Model | Header |
|----------|---------------|---------------|--------|
| Anthropic | User-provided (settings) | claude-sonnet-4-20250514 | x-anthropic-key |
| OpenAI | User-provided (settings) | gpt-4o | x-openai-key |
| Groq | Server env `GROQ_API_KEY` | openai/gpt-oss-20b | (none, server-side) |

## Key AI SDK Patterns

### Streaming Text
```typescript
import { streamText } from 'ai'

const result = streamText({
  model: aiModel,
  system: 'System prompt here',
  messages: modelMessages,
})

return result.toUIMessageStreamResponse()
```

### Tool Calling
```typescript
import { streamText, tool } from 'ai'
import { z } from 'zod'

const result = streamText({
  model: aiModel,
  system: systemPrompt,
  messages: modelMessages,
  tools: {
    getContacts: tool({
      description: 'Search for contacts in the connected CRM',
      parameters: z.object({
        query: z.string().describe('Search query'),
        limit: z.number().optional().describe('Max results'),
      }),
      execute: async ({ query, limit }) => {
        // Call backend API to search contacts
        return { contacts: [] }
      },
    }),
  },
})
```

### Client-Side Chat Hook
```tsx
import { useChat } from '@ai-sdk/react'

function ChatComponent() {
  const { messages, input, handleInputChange, handleSubmit, isLoading } = useChat({
    api: '/api/chat',
    headers: {
      'x-anthropic-key': apiKey,
    },
    body: {
      provider: 'anthropic',
      model: 'claude-sonnet-4-20250514',
    },
  })
}
```

### Object Generation
```typescript
import { generateObject } from 'ai'
import { z } from 'zod'

const result = await generateObject({
  model: aiModel,
  schema: z.object({
    name: z.string(),
    category: z.enum(['contact', 'data', 'tagging']),
  }),
  prompt: 'Generate a helper configuration...',
})
```

## Helper System Integration

The AI assistant can help users with their CRM data and Helpers. Key integration points:

- **Helper Configuration**: AI can suggest helper configs based on user descriptions
- **Data Analysis**: AI can analyze CRM contact data, deals, and pipelines
- **Troubleshooting**: AI can help debug CRM integration issues
- **Natural Language Queries**: The data explorer has a natural language query bar

Supported CRM platforms: Keap, GoHighLevel, ActiveCampaign, Ontraport, HubSpot.

## File Locations

- Chat API route: `src/app/api/chat/route.ts`
- Chat panel component: `src/components/ai-chat-panel.tsx`
- AI settings store: `src/lib/stores/ai-settings-store.ts`
- Data query routes: `src/app/api/data/query/route.ts`, `src/app/api/data/catalog/route.ts`
- NL query bar: `src/components/data-explorer/nl-query-bar.tsx`

## Dev & Test

```bash
cd /Users/nickkulavic/Projects/myfusionhelper.ai
npm run web       # Start dev server (port 3001)
npm run lint
npm run type-check
```
