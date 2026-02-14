import { apiClient } from './client'

export interface EmailTemplate {
  id: string
  name: string
  category: string
  subject: string
  body: string
  isStarred: boolean
  usageCount: number
  createdAt: string
  updatedAt: string
}

export interface Email {
  id: string
  subject: string
  body: string
  to: string
  status: 'draft' | 'scheduled' | 'sent' | 'failed'
  sentAt: string | null
  scheduledAt: string | null
  opens: number
  clicks: number
  templateId?: string
  createdAt: string
}

export interface CreateTemplateInput {
  name: string
  category: string
  subject: string
  body: string
}

export interface UpdateTemplateInput {
  name?: string
  category?: string
  subject?: string
  body?: string
  isStarred?: boolean
}

export interface SendEmailInput {
  to: string
  subject: string
  body: string
  templateId?: string
  scheduledAt?: string
}

// TODO: Connect to real backend endpoints when available
// For now, these will return mock data via the API client's error handling

export const emailsApi = {
  // Emails
  listEmails: () => apiClient.get<Email[]>('/emails'),

  getEmail: (id: string) => apiClient.get<Email>(`/emails/${id}`),

  sendEmail: (input: SendEmailInput) =>
    apiClient.post<Email>('/emails', input),

  deleteEmail: (id: string) => apiClient.delete<void>(`/emails/${id}`),

  // Templates
  listTemplates: () => apiClient.get<EmailTemplate[]>('/emails/templates'),

  getTemplate: (id: string) =>
    apiClient.get<EmailTemplate>(`/emails/templates/${id}`),

  createTemplate: (input: CreateTemplateInput) =>
    apiClient.post<EmailTemplate>('/emails/templates', input),

  updateTemplate: (id: string, input: UpdateTemplateInput) =>
    apiClient.put<EmailTemplate>(`/emails/templates/${id}`, input),

  deleteTemplate: (id: string) =>
    apiClient.delete<void>(`/emails/templates/${id}`),
}
