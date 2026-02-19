import {
  User,
  Building,
  Shield,
  Users,
  Key,
  Sparkles,
  CreditCard,
  Bell,
  type LucideIcon,
} from 'lucide-react'

export interface SettingsTab {
  id: string
  name: string
  icon: LucideIcon
}

export const settingsTabs: SettingsTab[] = [
  { id: 'profile', name: 'Profile', icon: User },
  { id: 'account', name: 'Account', icon: Building },
  { id: 'security', name: 'Security', icon: Shield },
  { id: 'team', name: 'Team', icon: Users },
  { id: 'api-keys', name: 'API Keys', icon: Key },
  { id: 'ai', name: 'AI Assistant', icon: Sparkles },
  { id: 'billing', name: 'Billing', icon: CreditCard },
  { id: 'notifications', name: 'Notifications', icon: Bell },
]
