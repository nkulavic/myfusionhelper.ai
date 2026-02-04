'use client'

import { useState } from 'react'
import { Plus, CheckCircle, XCircle, AlertCircle, ExternalLink, Settings, Trash2 } from 'lucide-react'
import { cn } from '@/lib/utils'

const platforms = [
  {
    id: 'keap',
    name: 'Keap',
    description: 'Formerly Infusionsoft',
    logo: '/platforms/keap.svg',
    status: 'available',
  },
  {
    id: 'gohighlevel',
    name: 'GoHighLevel',
    description: 'All-in-one marketing platform',
    logo: '/platforms/ghl.svg',
    status: 'available',
  },
  {
    id: 'activecampaign',
    name: 'ActiveCampaign',
    description: 'Email marketing & automation',
    logo: '/platforms/activecampaign.svg',
    status: 'coming_soon',
  },
  {
    id: 'ontraport',
    name: 'Ontraport',
    description: 'Business automation software',
    logo: '/platforms/ontraport.svg',
    status: 'coming_soon',
  },
]

const connections = [
  {
    id: '1',
    platform: 'keap',
    name: 'Production Keap',
    status: 'active',
    lastSync: '2 minutes ago',
    helpersCount: 12,
  },
  {
    id: '2',
    platform: 'keap',
    name: 'Sandbox Keap',
    status: 'error',
    lastSync: '1 hour ago',
    helpersCount: 3,
    error: 'Token expired',
  },
]

export default function ConnectionsPage() {
  const [showAddModal, setShowAddModal] = useState(false)

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'active':
        return <CheckCircle className="h-5 w-5 text-green-500" />
      case 'error':
        return <XCircle className="h-5 w-5 text-red-500" />
      default:
        return <AlertCircle className="h-5 w-5 text-yellow-500" />
    }
  }

  const getPlatformInfo = (platformId: string) => {
    return platforms.find((p) => p.id === platformId)
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Connections</h1>
          <p className="text-muted-foreground">Connect and manage your CRM platforms</p>
        </div>
        <button
          onClick={() => setShowAddModal(true)}
          className="inline-flex items-center gap-2 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
        >
          <Plus className="h-4 w-4" />
          Add Connection
        </button>
      </div>

      {/* Active Connections */}
      <div>
        <h2 className="mb-4 text-lg font-semibold">Active Connections</h2>
        {connections.length > 0 ? (
          <div className="space-y-4">
            {connections.map((connection) => {
              const platform = getPlatformInfo(connection.platform)
              return (
                <div
                  key={connection.id}
                  className="flex items-center justify-between rounded-lg border bg-card p-4"
                >
                  <div className="flex items-center gap-4">
                    <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-muted text-2xl font-bold">
                      {platform?.name.charAt(0)}
                    </div>
                    <div>
                      <div className="flex items-center gap-2">
                        <h3 className="font-semibold">{connection.name}</h3>
                        {getStatusIcon(connection.status)}
                      </div>
                      <div className="flex items-center gap-4 text-sm text-muted-foreground">
                        <span>{platform?.name}</span>
                        <span>•</span>
                        <span>{connection.helpersCount} helpers</span>
                        <span>•</span>
                        <span>Last sync: {connection.lastSync}</span>
                      </div>
                      {connection.error && (
                        <p className="mt-1 text-sm text-red-500">{connection.error}</p>
                      )}
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <button className="rounded-md p-2 hover:bg-accent">
                      <Settings className="h-4 w-4 text-muted-foreground" />
                    </button>
                    <button className="rounded-md p-2 hover:bg-accent">
                      <Trash2 className="h-4 w-4 text-muted-foreground" />
                    </button>
                  </div>
                </div>
              )
            })}
          </div>
        ) : (
          <div className="flex flex-col items-center justify-center rounded-lg border border-dashed py-12 text-center">
            <Plus className="mb-4 h-12 w-12 text-muted-foreground/50" />
            <h3 className="mb-1 font-semibold">No connections yet</h3>
            <p className="mb-4 text-sm text-muted-foreground">
              Connect your first CRM platform to start automating
            </p>
            <button
              onClick={() => setShowAddModal(true)}
              className="inline-flex items-center gap-2 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
            >
              <Plus className="h-4 w-4" />
              Add Connection
            </button>
          </div>
        )}
      </div>

      {/* Available Platforms */}
      <div>
        <h2 className="mb-4 text-lg font-semibold">Available Platforms</h2>
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {platforms.map((platform) => (
            <div
              key={platform.id}
              className={cn(
                'relative rounded-lg border bg-card p-4',
                platform.status === 'coming_soon' && 'opacity-60'
              )}
            >
              {platform.status === 'coming_soon' && (
                <span className="absolute right-2 top-2 rounded-full bg-muted px-2 py-0.5 text-xs font-medium">
                  Coming Soon
                </span>
              )}
              <div className="mb-3 flex h-12 w-12 items-center justify-center rounded-lg bg-muted text-2xl font-bold">
                {platform.name.charAt(0)}
              </div>
              <h3 className="font-semibold">{platform.name}</h3>
              <p className="mb-3 text-sm text-muted-foreground">{platform.description}</p>
              {platform.status === 'available' ? (
                <button className="inline-flex w-full items-center justify-center gap-2 rounded-md border border-input bg-background px-4 py-2 text-sm font-medium hover:bg-accent">
                  <ExternalLink className="h-4 w-4" />
                  Connect
                </button>
              ) : (
                <button
                  disabled
                  className="inline-flex w-full items-center justify-center gap-2 rounded-md border border-input bg-background px-4 py-2 text-sm font-medium opacity-50"
                >
                  Coming Soon
                </button>
              )}
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
