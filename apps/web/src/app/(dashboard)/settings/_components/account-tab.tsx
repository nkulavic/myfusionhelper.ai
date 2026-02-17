'use client'

import { useState } from 'react'
import { Copy, Loader2 } from 'lucide-react'
import { useWorkspaceStore } from '@/lib/stores/workspace-store'
import { useUpdateAccount } from '@/lib/hooks/use-settings'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { toast } from 'sonner'

export function AccountTab() {
  const { currentAccount } = useWorkspaceStore()
  const updateAccount = useUpdateAccount()
  const [accountName, setAccountName] = useState(currentAccount?.name || '')
  const [company, setCompany] = useState(currentAccount?.company || '')

  const handleSave = () => {
    if (!currentAccount) return
    updateAccount.mutate({
      accountId: currentAccount.accountId,
      input: { name: accountName, company },
    })
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Workspace Details</CardTitle>
          <CardDescription>Manage your workspace name and company info</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="workspace-name">Workspace Name</Label>
            <Input
              id="workspace-name"
              value={accountName}
              onChange={(e) => setAccountName(e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="workspace-company">Company</Label>
            <Input
              id="workspace-company"
              value={company}
              onChange={(e) => setCompany(e.target.value)}
            />
          </div>
          {currentAccount && (
            <div className="space-y-2">
              <Label>Workspace ID</Label>
              <div className="flex items-center gap-2">
                <Input
                  value={currentAccount.accountId}
                  readOnly
                  className="bg-muted font-mono text-muted-foreground"
                />
                <Button
                  variant="outline"
                  size="icon"
                  onClick={() => {
                    navigator.clipboard.writeText(currentAccount.accountId)
                    toast.success('Workspace ID copied')
                  }}
                >
                  <Copy className="h-4 w-4" />
                </Button>
              </div>
            </div>
          )}
        </CardContent>
        <CardFooter className="justify-end">
          <Button onClick={handleSave} disabled={updateAccount.isPending}>
            {updateAccount.isPending && <Loader2 className="h-4 w-4 animate-spin" />}
            Save Changes
          </Button>
        </CardFooter>
      </Card>

      <Card className="border-destructive/30">
        <CardHeader>
          <CardTitle className="text-lg text-destructive">Danger Zone</CardTitle>
          <CardDescription>Irreversible actions that affect your entire workspace</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-between rounded-md border border-destructive/30 p-4">
            <div>
              <p className="text-sm font-medium text-destructive">Delete Workspace</p>
              <p className="text-xs text-muted-foreground">
                Permanently delete this workspace and all associated data
              </p>
            </div>
            <div className="flex items-center gap-2">
              <Button variant="destructive" size="sm" disabled>
                Delete
              </Button>
              <span className="text-xs text-muted-foreground">Coming soon</span>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
