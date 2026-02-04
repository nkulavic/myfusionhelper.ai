'use client'

import { useState } from 'react'
import { Search, Filter, CheckCircle, XCircle, Clock, RefreshCw } from 'lucide-react'
import { cn } from '@/lib/utils'

const executions = [
  {
    id: '1',
    helper: 'Tag It',
    contact: 'john@example.com',
    status: 'completed',
    duration: '125ms',
    timestamp: '2 minutes ago',
  },
  {
    id: '2',
    helper: 'Copy It',
    contact: 'jane@example.com',
    status: 'completed',
    duration: '89ms',
    timestamp: '5 minutes ago',
  },
  {
    id: '3',
    helper: 'Google Sheet It',
    contact: 'bob@example.com',
    status: 'failed',
    duration: '2.3s',
    timestamp: '10 minutes ago',
    error: 'Rate limit exceeded',
  },
  {
    id: '4',
    helper: 'Notify Me',
    contact: 'alice@example.com',
    status: 'running',
    duration: '-',
    timestamp: 'Just now',
  },
  {
    id: '5',
    helper: 'Date Calculator',
    contact: 'charlie@example.com',
    status: 'completed',
    duration: '45ms',
    timestamp: '15 minutes ago',
  },
]

export default function ExecutionsPage() {
  const [searchQuery, setSearchQuery] = useState('')
  const [statusFilter, setStatusFilter] = useState('all')

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'completed':
        return <CheckCircle className="h-4 w-4 text-green-500" />
      case 'failed':
        return <XCircle className="h-4 w-4 text-red-500" />
      case 'running':
        return <RefreshCw className="h-4 w-4 animate-spin text-blue-500" />
      default:
        return <Clock className="h-4 w-4 text-yellow-500" />
    }
  }

  const filteredExecutions = executions.filter((execution) => {
    const matchesSearch =
      execution.helper.toLowerCase().includes(searchQuery.toLowerCase()) ||
      execution.contact.toLowerCase().includes(searchQuery.toLowerCase())
    const matchesStatus = statusFilter === 'all' || execution.status === statusFilter
    return matchesSearch && matchesStatus
  })

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold">Executions</h1>
        <p className="text-muted-foreground">View and monitor helper execution history</p>
      </div>

      {/* Stats */}
      <div className="grid gap-4 sm:grid-cols-4">
        <div className="rounded-lg border bg-card p-4">
          <p className="text-sm text-muted-foreground">Total Today</p>
          <p className="text-2xl font-bold">1,247</p>
        </div>
        <div className="rounded-lg border bg-card p-4">
          <p className="text-sm text-muted-foreground">Success Rate</p>
          <p className="text-2xl font-bold text-green-500">98.5%</p>
        </div>
        <div className="rounded-lg border bg-card p-4">
          <p className="text-sm text-muted-foreground">Avg Duration</p>
          <p className="text-2xl font-bold">142ms</p>
        </div>
        <div className="rounded-lg border bg-card p-4">
          <p className="text-sm text-muted-foreground">Failed</p>
          <p className="text-2xl font-bold text-red-500">19</p>
        </div>
      </div>

      {/* Filters */}
      <div className="flex gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <input
            type="text"
            placeholder="Search by helper or contact..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="h-10 w-full rounded-md border border-input bg-background pl-10 pr-4 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          />
        </div>
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          className="h-10 rounded-md border border-input bg-background px-3 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          <option value="all">All Status</option>
          <option value="completed">Completed</option>
          <option value="failed">Failed</option>
          <option value="running">Running</option>
        </select>
        <button className="inline-flex items-center gap-2 rounded-md border border-input bg-background px-4 py-2 text-sm font-medium hover:bg-accent">
          <Filter className="h-4 w-4" />
          More Filters
        </button>
      </div>

      {/* Table */}
      <div className="rounded-lg border">
        <table className="w-full">
          <thead>
            <tr className="border-b bg-muted/50">
              <th className="p-4 text-left text-sm font-medium text-muted-foreground">Status</th>
              <th className="p-4 text-left text-sm font-medium text-muted-foreground">Helper</th>
              <th className="p-4 text-left text-sm font-medium text-muted-foreground">Contact</th>
              <th className="p-4 text-left text-sm font-medium text-muted-foreground">Duration</th>
              <th className="p-4 text-left text-sm font-medium text-muted-foreground">Time</th>
            </tr>
          </thead>
          <tbody>
            {filteredExecutions.map((execution) => (
              <tr
                key={execution.id}
                className="border-b last:border-0 hover:bg-muted/50 cursor-pointer"
              >
                <td className="p-4">
                  <div className="flex items-center gap-2">
                    {getStatusIcon(execution.status)}
                    <span
                      className={cn(
                        'text-sm capitalize',
                        execution.status === 'completed' && 'text-green-600',
                        execution.status === 'failed' && 'text-red-600',
                        execution.status === 'running' && 'text-blue-600'
                      )}
                    >
                      {execution.status}
                    </span>
                  </div>
                </td>
                <td className="p-4 font-medium">{execution.helper}</td>
                <td className="p-4 text-muted-foreground">{execution.contact}</td>
                <td className="p-4 font-mono text-sm">{execution.duration}</td>
                <td className="p-4 text-sm text-muted-foreground">{execution.timestamp}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
