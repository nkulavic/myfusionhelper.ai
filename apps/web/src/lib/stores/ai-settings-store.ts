import { create } from 'zustand'
import { persist } from 'zustand/middleware'

export type AIProvider = 'groq' | 'anthropic' | 'openai'

export interface ProviderModels {
  groq: string[]
  anthropic: string[]
  openai: string[]
}

export const providerModels: ProviderModels = {
  groq: ['openai/gpt-oss-20b', 'openai/gpt-oss-120b', 'qwen/qwen3-32b', 'llama-3.3-70b-versatile'],
  anthropic: ['claude-sonnet-4-20250514', 'claude-opus-4-0-20250514'],
  openai: ['gpt-4o', 'gpt-4o-mini'],
}

export const providerLabels: Record<AIProvider, string> = {
  groq: 'Groq (Free)',
  anthropic: 'Claude (BYOK)',
  openai: 'OpenAI (BYOK)',
}

export const modelLabels: Record<string, string> = {
  'openai/gpt-oss-20b': 'GPT OSS 20B (Fastest)',
  'openai/gpt-oss-120b': 'GPT OSS 120B',
  'qwen/qwen3-32b': 'Qwen3 32B',
  'llama-3.3-70b-versatile': 'Llama 3.3 70B',
  'claude-sonnet-4-20250514': 'Claude Sonnet 4',
  'claude-opus-4-0-20250514': 'Claude Opus 4',
  'gpt-4o': 'GPT-4o',
  'gpt-4o-mini': 'GPT-4o Mini',
}

interface AISettingsState {
  anthropicApiKey: string | null
  openaiApiKey: string | null
  preferredProvider: AIProvider
  preferredModel: string
  setAnthropicKey: (key: string | null) => void
  setOpenAIKey: (key: string | null) => void
  setPreferredProvider: (provider: AIProvider) => void
  setPreferredModel: (model: string) => void
}

export const useAISettingsStore = create<AISettingsState>()(
  persist(
    (set) => ({
      anthropicApiKey: null,
      openaiApiKey: null,
      preferredProvider: 'groq',
      preferredModel: 'openai/gpt-oss-20b',
      setAnthropicKey: (key) => set({ anthropicApiKey: key }),
      setOpenAIKey: (key) => set({ openaiApiKey: key }),
      setPreferredProvider: (provider) =>
        set({
          preferredProvider: provider,
          preferredModel: providerModels[provider][0],
        }),
      setPreferredModel: (model) => set({ preferredModel: model }),
    }),
    {
      name: 'mfh-ai-settings',
      partialize: (state) => ({
        anthropicApiKey: state.anthropicApiKey,
        openaiApiKey: state.openaiApiKey,
        preferredProvider: state.preferredProvider,
        preferredModel: state.preferredModel,
      }),
    }
  )
)
