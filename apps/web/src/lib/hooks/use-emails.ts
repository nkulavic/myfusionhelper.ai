import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  emailsApi,
  type Email,
  type EmailTemplate,
  type CreateTemplateInput,
  type UpdateTemplateInput,
  type SendEmailInput,
} from '@/lib/api/emails'

// ---------- Mock data (until backend endpoints exist) ----------

const mockEmails: Email[] = [
  {
    id: 'em_001',
    subject: 'Welcome to our community!',
    body: 'Hi {{first_name}}, we are thrilled to have you on board...',
    to: 'New Leads Segment',
    status: 'sent',
    sentAt: '2026-02-04T14:00:00Z',
    scheduledAt: null,
    opens: 847,
    clicks: 234,
    createdAt: '2026-02-04T13:00:00Z',
  },
  {
    id: 'em_002',
    subject: 'Your weekly automation report',
    body: 'Here is your weekly summary of helper activity...',
    to: 'All Contacts',
    status: 'scheduled',
    sentAt: null,
    scheduledAt: '2026-02-07T09:00:00Z',
    opens: 0,
    clicks: 0,
    createdAt: '2026-02-03T10:00:00Z',
  },
  {
    id: 'em_003',
    subject: 'Special offer for loyal customers',
    body: 'As a valued customer, we have a special deal for you...',
    to: 'VIP Segment',
    status: 'draft',
    sentAt: null,
    scheduledAt: null,
    opens: 0,
    clicks: 0,
    createdAt: '2026-02-01T08:00:00Z',
  },
]

const mockTemplates: EmailTemplate[] = [
  {
    id: 'tpl_001',
    name: 'Welcome Email',
    category: 'onboarding',
    subject: 'Welcome to {{company}}, {{first_name}}!',
    body: 'Hi {{first_name}},\n\nWe are thrilled to have you on board. Here are a few things to get started...',
    isStarred: true,
    usageCount: 1247,
    createdAt: '2025-12-15T00:00:00Z',
    updatedAt: '2026-01-20T00:00:00Z',
  },
  {
    id: 'tpl_002',
    name: 'Follow-up After Demo',
    category: 'sales',
    subject: 'Great connecting today, {{first_name}}',
    body: 'Thank you for taking the time to see our demo today. I wanted to follow up on...',
    isStarred: true,
    usageCount: 856,
    createdAt: '2026-01-05T00:00:00Z',
    updatedAt: '2026-01-05T00:00:00Z',
  },
  {
    id: 'tpl_003',
    name: 'Payment Reminder',
    category: 'billing',
    subject: 'Friendly reminder: Invoice due soon',
    body: 'Hi {{first_name}}, this is a quick reminder that your invoice is due on...',
    isStarred: false,
    usageCount: 423,
    createdAt: '2026-01-10T00:00:00Z',
    updatedAt: '2026-01-10T00:00:00Z',
  },
  {
    id: 'tpl_004',
    name: 'Re-engagement Campaign',
    category: 'marketing',
    subject: 'We miss you, {{first_name}}!',
    body: "It's been a while since we last connected. We wanted to reach out and share...",
    isStarred: false,
    usageCount: 312,
    createdAt: '2026-01-20T00:00:00Z',
    updatedAt: '2026-01-20T00:00:00Z',
  },
  {
    id: 'tpl_005',
    name: 'Monthly Newsletter',
    category: 'marketing',
    subject: 'Your {{company}} Monthly Update',
    body: "Here's what happened this month and what's coming next...",
    isStarred: true,
    usageCount: 2100,
    createdAt: '2025-11-01T00:00:00Z',
    updatedAt: '2026-02-01T00:00:00Z',
  },
  {
    id: 'tpl_006',
    name: 'Webinar Invitation',
    category: 'events',
    subject: "You're invited: {{event_name}}",
    body: 'Join us for an exclusive webinar on automation strategies...',
    isStarred: false,
    usageCount: 645,
    createdAt: '2026-02-01T00:00:00Z',
    updatedAt: '2026-02-01T00:00:00Z',
  },
]

// Helper to simulate API delay
function mockDelay<T>(data: T, ms = 300): Promise<{ data: T }> {
  return new Promise((resolve) => setTimeout(() => resolve({ data }), ms))
}

// ---------- Email hooks ----------

export function useEmails() {
  return useQuery({
    queryKey: ['emails'],
    queryFn: async () => {
      try {
        const res = await emailsApi.listEmails()
        return (res.data ?? []) as Email[]
      } catch {
        // Fall back to mock data when endpoint doesn't exist
        const res = await mockDelay(mockEmails)
        return res.data
      }
    },
  })
}

export function useSendEmail() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (input: SendEmailInput) => {
      try {
        return await emailsApi.sendEmail(input)
      } catch {
        // Mock: create a new email locally
        const newEmail: Email = {
          id: `em_${Date.now()}`,
          subject: input.subject,
          body: input.body,
          to: input.to,
          status: input.scheduledAt ? 'scheduled' : 'sent',
          sentAt: input.scheduledAt ? null : new Date().toISOString(),
          scheduledAt: input.scheduledAt || null,
          opens: 0,
          clicks: 0,
          templateId: input.templateId,
          createdAt: new Date().toISOString(),
        }
        return { data: newEmail }
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['emails'] })
    },
  })
}

export function useDeleteEmail() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (id: string) => {
      try {
        return await emailsApi.deleteEmail(id)
      } catch {
        await mockDelay(undefined, 200)
        return { data: undefined }
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['emails'] })
    },
  })
}

// ---------- Template hooks ----------

export function useEmailTemplates() {
  return useQuery({
    queryKey: ['email-templates'],
    queryFn: async () => {
      try {
        const res = await emailsApi.listTemplates()
        return (res.data ?? []) as EmailTemplate[]
      } catch {
        const res = await mockDelay(mockTemplates)
        return res.data
      }
    },
  })
}

export function useEmailTemplate(id: string) {
  return useQuery({
    queryKey: ['email-templates', id],
    queryFn: async () => {
      try {
        const res = await emailsApi.getTemplate(id)
        return res.data
      } catch {
        const template = mockTemplates.find((t) => t.id === id)
        if (!template) throw new Error('Template not found')
        return template
      }
    },
    enabled: !!id,
  })
}

export function useCreateTemplate() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (input: CreateTemplateInput) => {
      try {
        return await emailsApi.createTemplate(input)
      } catch {
        const newTemplate: EmailTemplate = {
          id: `tpl_${Date.now()}`,
          ...input,
          isStarred: false,
          usageCount: 0,
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
        }
        return { data: newTemplate }
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['email-templates'] })
    },
  })
}

export function useUpdateTemplate() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, input }: { id: string; input: UpdateTemplateInput }) => {
      try {
        return await emailsApi.updateTemplate(id, input)
      } catch {
        await mockDelay(undefined, 200)
        return { data: undefined }
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['email-templates'] })
    },
  })
}

export function useDeleteTemplate() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (id: string) => {
      try {
        return await emailsApi.deleteTemplate(id)
      } catch {
        await mockDelay(undefined, 200)
        return { data: undefined }
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['email-templates'] })
    },
  })
}
