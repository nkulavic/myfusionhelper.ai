'use client'

import { useState } from 'react'
import { Eye, Key, Loader2, Plus, Shield, Trash2, Users, XCircle } from 'lucide-react'
import { useAuthStore } from '@/lib/stores/auth-store'
import { useWorkspaceStore } from '@/lib/stores/workspace-store'
import {
  useTeamMembers,
  useInviteTeamMember,
  useUpdateTeamMember,
  useRemoveTeamMember,
} from '@/lib/hooks/use-settings'
import { usePlanLimits } from '@/lib/hooks/use-plan-limits'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'

export function TeamTab() {
  const { user } = useAuthStore()
  const { currentAccount } = useWorkspaceStore()
  const { data: members, isLoading, isError } = useTeamMembers(currentAccount?.accountId || '')
  const inviteMember = useInviteTeamMember()
  const updateMember = useUpdateTeamMember()
  const removeMember = useRemoveTeamMember()
  const { canCreate: canCreateResource, getUsage, getLimit } = usePlanLimits()
  const [showInvite, setShowInvite] = useState(false)
  const [inviteEmail, setInviteEmail] = useState('')
  const [inviteRole, setInviteRole] = useState<'admin' | 'member' | 'viewer'>('member')
  const [confirmRemove, setConfirmRemove] = useState<string | null>(null)
  const atTeamLimit = !canCreateResource('teamMembers')

  const handleInvite = () => {
    if (!currentAccount || !inviteEmail.trim()) return
    inviteMember.mutate(
      { accountId: currentAccount.accountId, input: { email: inviteEmail, role: inviteRole } },
      {
        onSuccess: () => {
          setInviteEmail('')
          setShowInvite(false)
        },
      }
    )
  }

  const handleRoleChange = (userId: string, role: 'admin' | 'member' | 'viewer') => {
    if (!currentAccount) return
    updateMember.mutate({
      accountId: currentAccount.accountId,
      userId,
      input: { role },
    })
  }

  const handleRemove = (userId: string) => {
    if (!currentAccount) return
    removeMember.mutate(
      { accountId: currentAccount.accountId, userId },
      { onSuccess: () => setConfirmRemove(null) }
    )
  }

  const userInitials =
    user?.name
      ?.split(' ')
      .map((n) => n[0])
      .join('')
      .toUpperCase()
      .slice(0, 2) || 'U'

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="text-lg">Team Members</CardTitle>
              <CardDescription>Manage who has access to this workspace</CardDescription>
            </div>
            <div className="flex items-center gap-2">
              <span className="text-xs text-muted-foreground">
                {getUsage('teamMembers')} / {getLimit('teamMembers')} members
              </span>
              <Button onClick={() => setShowInvite(!showInvite)} disabled={atTeamLimit}>
                <Plus className="h-4 w-4" />
                Invite Member
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          {inviteMember.isError && (
            <div className="flex items-center gap-2 rounded-md border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive">
              <XCircle className="h-4 w-4 shrink-0" />
              {(inviteMember.error as Error)?.message || 'Failed to send invite'}
            </div>
          )}

          {showInvite && (
            <div className="flex items-center gap-2 rounded-md border p-4">
              <Input
                type="email"
                value={inviteEmail}
                onChange={(e) => setInviteEmail(e.target.value)}
                placeholder="colleague@company.com"
                className="flex-1"
              />
              <select
                value={inviteRole}
                onChange={(e) => setInviteRole(e.target.value as 'admin' | 'member' | 'viewer')}
                className="flex h-10 rounded-md border border-input bg-background px-3 py-2 text-sm"
              >
                <option value="admin">Admin</option>
                <option value="member">Member</option>
                <option value="viewer">Viewer</option>
              </select>
              <Button
                onClick={handleInvite}
                disabled={inviteMember.isPending || !inviteEmail.trim()}
              >
                {inviteMember.isPending && <Loader2 className="h-4 w-4 animate-spin" />}
                Send Invite
              </Button>
              <Button variant="outline" onClick={() => setShowInvite(false)}>
                Cancel
              </Button>
            </div>
          )}

          {/* Current user always shows */}
          <div className="flex items-center justify-between rounded-md border p-4">
            <div className="flex items-center gap-3">
              <Avatar>
                <AvatarFallback className="bg-primary text-sm font-medium text-primary-foreground">
                  {userInitials}
                </AvatarFallback>
              </Avatar>
              <div>
                <p className="text-sm font-medium">{user?.name || 'You'}</p>
                <p className="text-xs text-muted-foreground">{user?.email || ''}</p>
              </div>
            </div>
            <Badge>Owner</Badge>
          </div>

          {isLoading ? (
            <div className="space-y-3">
              {[1, 2].map((i) => (
                <Skeleton key={i} className="h-16 w-full" />
              ))}
            </div>
          ) : isError ? (
            <div className="py-4 text-center text-sm text-muted-foreground">
              Failed to load team members. Please try again.
            </div>
          ) : members && members.length > 0 ? (
            members.map(
              (member: {
                userId: string
                name: string
                email: string
                role: string
                status: string
              }) => (
                <div
                  key={member.userId}
                  className="flex items-center justify-between rounded-md border p-4"
                >
                  <div className="flex items-center gap-3">
                    <Avatar>
                      <AvatarFallback>
                        {member.name
                          ?.split(' ')
                          .map((n: string) => n[0])
                          .join('')
                          .toUpperCase()
                          .slice(0, 2) || '??'}
                      </AvatarFallback>
                    </Avatar>
                    <div>
                      <p className="text-sm font-medium">
                        {member.name}
                        {member.status === 'Pending' && (
                          <span className="ml-2 text-xs text-muted-foreground">(pending)</span>
                        )}
                      </p>
                      <p className="text-xs text-muted-foreground">{member.email}</p>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    {member.role !== 'Owner' ? (
                      <>
                        <select
                          value={member.role.toLowerCase()}
                          onChange={(e) =>
                            handleRoleChange(
                              member.userId,
                              e.target.value as 'admin' | 'member' | 'viewer'
                            )
                          }
                          disabled={updateMember.isPending}
                          className="flex h-8 rounded-md border border-input bg-background px-2 py-1 text-xs"
                        >
                          <option value="admin">Admin</option>
                          <option value="member">Member</option>
                          <option value="viewer">Viewer</option>
                        </select>
                        {confirmRemove === member.userId ? (
                          <div className="flex items-center gap-1">
                            <Button
                              variant="destructive"
                              size="sm"
                              onClick={() => handleRemove(member.userId)}
                              disabled={removeMember.isPending}
                            >
                              {removeMember.isPending ? (
                                <Loader2 className="h-3 w-3 animate-spin" />
                              ) : (
                                'Confirm'
                              )}
                            </Button>
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => setConfirmRemove(null)}
                            >
                              Cancel
                            </Button>
                          </div>
                        ) : (
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-8 w-8 text-muted-foreground hover:text-destructive"
                            onClick={() => setConfirmRemove(member.userId)}
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        )}
                      </>
                    ) : (
                      <Badge>Owner</Badge>
                    )}
                  </div>
                </div>
              )
            )
          ) : null}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Roles & Permissions</CardTitle>
          <CardDescription>Understanding access levels in your workspace</CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          {[
            {
              role: 'Owner',
              variant: 'default' as const,
              icon: Shield,
              desc: 'Full access to everything including billing, team management, and workspace deletion',
            },
            {
              role: 'Admin',
              variant: 'info' as const,
              icon: Key,
              desc: 'Can manage helpers, connections, team members, and API keys. Cannot access billing or delete workspace.',
            },
            {
              role: 'Member',
              variant: 'success' as const,
              icon: Users,
              desc: 'Can manage helpers and execute them. Cannot manage connections, team, or API keys.',
            },
            {
              role: 'Viewer',
              variant: 'secondary' as const,
              icon: Eye,
              desc: 'Read-only access to helpers, executions, and analytics. Cannot make changes.',
            },
          ].map((item) => (
            <div key={item.role} className="flex items-start gap-3">
              <Badge variant={item.variant} className="mt-0.5 w-16 justify-center">
                {item.role}
              </Badge>
              <p className="flex-1 text-sm text-muted-foreground">{item.desc}</p>
            </div>
          ))}
        </CardContent>
      </Card>
    </div>
  )
}
