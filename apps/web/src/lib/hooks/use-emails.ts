import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  emailsApi,
  type Email,
  type EmailTemplate,
  type CreateTemplateInput,
  type UpdateTemplateInput,
  type SendEmailInput,
} from '@/lib/api/emails'

// ---------- Email hooks ----------

export function useEmails() {
  return useQuery({
    queryKey: ['emails'],
    queryFn: async () => {
      const res = await emailsApi.listEmails()
      return (res.data ?? []) as Email[]
    },
  })
}

export function useSendEmail() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: SendEmailInput) => emailsApi.sendEmail(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['emails'] })
    },
  })
}

export function useDeleteEmail() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => emailsApi.deleteEmail(id),
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
      const res = await emailsApi.listTemplates()
      return (res.data ?? []) as EmailTemplate[]
    },
  })
}

export function useEmailTemplate(id: string) {
  return useQuery({
    queryKey: ['email-templates', id],
    queryFn: async () => {
      const res = await emailsApi.getTemplate(id)
      return res.data
    },
    enabled: !!id,
  })
}

export function useCreateTemplate() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateTemplateInput) => emailsApi.createTemplate(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['email-templates'] })
    },
  })
}

export function useUpdateTemplate() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: UpdateTemplateInput }) =>
      emailsApi.updateTemplate(id, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['email-templates'] })
    },
  })
}

export function useDeleteTemplate() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => emailsApi.deleteTemplate(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['email-templates'] })
    },
  })
}
