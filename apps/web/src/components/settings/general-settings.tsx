'use client'

import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Skeleton } from '@/components/ui/skeleton'
import { useTenant, useUpdateTenant } from '@/lib/hooks'
import { toast } from 'sonner'
import { Loader2 } from 'lucide-react'

export function GeneralSettings() {
  const { data: tenant, isLoading, error } = useTenant()
  const updateTenant = useUpdateTenant()
  const [tenantName, setTenantName] = useState<string | null>(null)
  const [defaultProvider, setDefaultProvider] = useState<string | null>(null)
  const [darkMode, setDarkMode] = useState<boolean | null>(null)
  const [hasChanges, setHasChanges] = useState(false)

  // Update local state when tenant data loads
  useEffect(() => {
    if (tenant) {
      setTenantName(tenant.name)
      setDefaultProvider(tenant.defaultProvider || 'auto')
      setDarkMode(tenant.settings?.darkMode ?? true)
    }
  }, [tenant])

  // Track changes
  useEffect(() => {
    if (tenant && tenantName !== null) {
      const nameChanged = tenantName !== tenant.name
      const providerChanged = defaultProvider !== (tenant.defaultProvider || 'auto')
      const darkModeChanged = darkMode !== (tenant.settings?.darkMode ?? true)
      setHasChanges(nameChanged || providerChanged || darkModeChanged)
    }
  }, [tenantName, defaultProvider, darkMode, tenant])

  const handleSave = async () => {
    try {
      await updateTenant.mutateAsync({
        name: tenantName!,
        defaultProvider: defaultProvider === 'auto' ? '' : defaultProvider!,
        settings: {
          darkMode: darkMode!,
        },
      })

      setHasChanges(false)
      toast.success('Settings saved', {
        description: 'Your preferences have been updated.',
      })
    } catch (err) {
      toast.error('Failed to save settings', {
        description: 'Please try again later.',
      })
    }
  }

  if (isLoading || tenantName === null) {
    return (
      <div className="space-y-6">
        <Card>
          <CardHeader>
            <Skeleton className="h-6 w-48" />
            <Skeleton className="h-4 w-64 mt-2" />
          </CardHeader>
          <CardContent className="space-y-4">
            <Skeleton className="h-10 w-full" />
            <Skeleton className="h-10 w-full" />
          </CardContent>
        </Card>
      </div>
    )
  }

  if (error) {
    return (
      <Card className="border-destructive">
        <CardHeader>
          <CardTitle className="text-destructive">Error Loading Settings</CardTitle>
          <CardDescription>
            Failed to load tenant settings. Please try again later.
          </CardDescription>
        </CardHeader>
      </Card>
    )
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>Tenant Information</CardTitle>
          <CardDescription>
            Configure your tenant settings and preferences
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="tenant-id">Tenant ID</Label>
            <Input
              id="tenant-id"
              value={tenant?.id || ''}
              disabled
              className="font-mono text-sm bg-muted"
            />
            <p className="text-xs text-muted-foreground">
              Your unique tenant identifier (read-only)
            </p>
          </div>

          <div className="space-y-2">
            <Label htmlFor="tenant-name">Tenant Name</Label>
            <Input
              id="tenant-name"
              value={tenantName!}
              onChange={(e) => setTenantName(e.target.value)}
              placeholder="Enter tenant name"
            />
          </div>

          <div className="space-y-2">
            <Label>Status</Label>
            <div className="flex items-center gap-2">
              <div className={`h-2 w-2 rounded-full ${tenant?.active ? 'bg-green-500' : 'bg-red-500'}`} />
              <span className="text-sm">{tenant?.active ? 'Active' : 'Inactive'}</span>
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="default-provider">Default Provider</Label>
            <Select value={defaultProvider!} onValueChange={(v) => setDefaultProvider(v)}>
              <SelectTrigger id="default-provider">
                <SelectValue placeholder="Select provider" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="auto">Auto (Best Available)</SelectItem>
                <SelectItem value="openai">OpenAI</SelectItem>
                <SelectItem value="claude">Claude</SelectItem>
                <SelectItem value="ollama">Ollama</SelectItem>
                <SelectItem value="local">Local</SelectItem>
              </SelectContent>
            </Select>
            <p className="text-xs text-muted-foreground">
              The default provider used when none is specified
            </p>
          </div>

          <div className="grid grid-cols-2 gap-4 pt-4 border-t">
            <div>
              <Label className="text-muted-foreground text-xs">Created</Label>
              <p className="text-sm">
                {tenant?.createdAt ? new Date(tenant.createdAt).toLocaleDateString() : '-'}
              </p>
            </div>
            <div>
              <Label className="text-muted-foreground text-xs">Last Updated</Label>
              <p className="text-sm">
                {tenant?.updatedAt ? new Date(tenant.updatedAt).toLocaleDateString() : '-'}
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Appearance</CardTitle>
          <CardDescription>
            Customize the look and feel of the dashboard
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label>Dark Mode</Label>
              <p className="text-xs text-muted-foreground">
                Use dark theme for the dashboard
              </p>
            </div>
            <Switch checked={darkMode!} onCheckedChange={(v) => setDarkMode(v)} />
          </div>
        </CardContent>
      </Card>

      <div className="flex justify-end">
        <Button onClick={handleSave} disabled={updateTenant.isPending || !hasChanges}>
          {updateTenant.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
          Save Changes
        </Button>
      </div>
    </div>
  )
}
