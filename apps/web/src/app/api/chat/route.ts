import { createGroq } from '@ai-sdk/groq'
import { createAnthropic } from '@ai-sdk/anthropic'
import { createOpenAI } from '@ai-sdk/openai'
import { streamText, convertToModelMessages } from 'ai'

const systemPrompt = `You are an AI assistant for MyFusion Helper, a CRM automation platform that helps businesses automate their customer relationship management workflows.

You can help users with:
- Understanding their CRM data and automations
- Configuring helpers (automation tasks) for their CRM
- Analyzing contact data, deals, and pipelines
- Creating reports and insights about their business
- Troubleshooting issues with their CRM integrations

Supported platforms: Keap, GoHighLevel, ActiveCampaign, Ontraport, HubSpot, and Stripe.

Be concise, helpful, and focused on CRM automation topics. When you don't know something specific about the user's data, let them know what information you'd need.`

export async function POST(req: Request) {
  const { messages, provider, model } = await req.json()

  let aiModel
  if (provider === 'anthropic') {
    const key = req.headers.get('x-anthropic-key')
    if (!key) {
      return new Response('Anthropic API key required. Configure it in Settings > AI Assistant.', {
        status: 401,
      })
    }
    const anthropic = createAnthropic({ apiKey: key })
    aiModel = anthropic(model || 'claude-sonnet-4-20250514')
  } else if (provider === 'openai') {
    const key = req.headers.get('x-openai-key')
    if (!key) {
      return new Response('OpenAI API key required. Configure it in Settings > AI Assistant.', {
        status: 401,
      })
    }
    const openai = createOpenAI({ apiKey: key })
    aiModel = openai(model || 'gpt-4o')
  } else {
    const groqKey = process.env.GROQ_API_KEY
    if (!groqKey) {
      return new Response('Groq API key not configured on the server.', { status: 500 })
    }
    const groq = createGroq({ apiKey: groqKey })
    aiModel = groq(model || 'openai/gpt-oss-20b')
  }

  const modelMessages = await convertToModelMessages(messages)

  const result = streamText({
    model: aiModel,
    system: systemPrompt,
    messages: modelMessages,
  })

  return result.toUIMessageStreamResponse()
}
