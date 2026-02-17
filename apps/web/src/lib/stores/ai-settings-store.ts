import { create } from 'zustand'
import { persist } from 'zustand/middleware'

export const groqModels = [
  'openai/gpt-oss-120b',
  'openai/gpt-oss-20b',
  'qwen/qwen3-32b',
  'llama-3.3-70b-versatile',
] as const

export const modelLabels: Record<string, string> = {
  'openai/gpt-oss-120b': 'GPT OSS 120B (Recommended)',
  'openai/gpt-oss-20b': 'GPT OSS 20B (Fastest)',
  'qwen/qwen3-32b': 'Qwen3 32B',
  'llama-3.3-70b-versatile': 'Llama 3.3 70B',
}

interface AISettingsState {
  preferredModel: string
  setPreferredModel: (model: string) => void
}

export const useAISettingsStore = create<AISettingsState>()(
  persist(
    (set) => ({
      preferredModel: 'openai/gpt-oss-120b',
      setPreferredModel: (model) => set({ preferredModel: model }),
    }),
    {
      name: 'mfh-ai-settings',
      partialize: (state) => ({
        preferredModel: state.preferredModel,
      }),
    }
  )
)
