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
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import { Skeleton } from '@/components/ui/skeleton'
import { toast } from 'sonner'
import { Loader2 } from 'lucide-react'
import { useTenant, useUpdateTenant } from '@/lib/hooks'

export function NotificationsSettings() {
  const { data: tenant, isLoading } = useTenant()
  const updateTenant = useUpdateTenant()

  // Initialize with null to detect when we have real values
  const [jobCompleted, setJobCompleted] = useState<boolean | null>(null)
  const [jobFailed, setJobFailed] = useState<boolean | null>(null)
  const [providerOffline, setProviderOffline] = useState<boolean | null>(null)
  const [usageThreshold, setUsageThreshold] = useState<boolean | null>(null)
  const [weeklySummary, setWeeklySummary] = useState<boolean | null>(null)
  const [marketingEmails, setMarketingEmails] = useState<boolean | null>(null)
  const [hasChanges, setHasChanges] = useState(false)

  // Load settings from tenant
  useEffect(() => {
    if (tenant?.settings?.notifications) {
      const n = tenant.settings.notifications
      setJobCompleted(n.jobCompleted)
      setJobFailed(n.jobFailed)
      setProviderOffline(n.providerOffline)
      setUsageThreshold(n.usageThreshold)
      setWeeklySummary(n.weeklySummary)
      setMarketingEmails(n.marketingEmails)
    }
  }, [tenant])

  // Track changes
  useEffect(() => {
    if (tenant?.settings?.notifications && jobCompleted !== null) {
      const n = tenant.settings.notifications
      const changed =
        jobCompleted !== n.jobCompleted ||
        jobFailed !== n.jobFailed ||
        providerOffline !== n.providerOffline ||
        usageThreshold !== n.usageThreshold ||
        weeklySummary !== n.weeklySummary ||
        marketingEmails !== n.marketingEmails
      setHasChanges(changed)
    }
  }, [jobCompleted, jobFailed, providerOffline, usageThreshold, weeklySummary, marketingEmails, tenant])

  const handleSave = async () => {
    try {
      await updateTenant.mutateAsync({
        settings: {
          notifications: {
            jobCompleted: jobCompleted!,
            jobFailed: jobFailed!,
            providerOffline: providerOffline!,
            usageThreshold: usageThreshold!,
            weeklySummary: weeklySummary!,
            marketingEmails: marketingEmails!,
          },
        },
      })

      setHasChanges(false)
      toast.success('Preferences saved', {
        description: 'Your notification settings have been updated.',
      })
    } catch (err) {
      toast.error('Failed to save settings', {
        description: 'Please try again later.',
      })
    }
  }

  if (isLoading || jobCompleted === null) {
    return (
      <div className="space-y-6">
        <Card>
          <CardHeader>
            <Skeleton className="h-6 w-48" />
            <Skeleton className="h-4 w-64 mt-2" />
          </CardHeader>
          <CardContent className="space-y-4">
            {[1, 2, 3, 4, 5].map((i) => (
              <div key={i} className="flex items-center justify-between">
                <Skeleton className="h-10 w-48" />
                <Skeleton className="h-6 w-12" />
              </div>
            ))}
          </CardContent>
        </Card>
      </div>
    )
  }

  const settings = [
    {
      id: 'job-completed',
      title: 'Job Completed',
      description: 'Notify when a job finishes processing',
      enabled: jobCompleted!,
      onChange: (v: boolean) => setJobCompleted(v),
    },
    {
      id: 'job-failed',
      title: 'Job Failed',
      description: 'Notify when a job fails to process',
      enabled: jobFailed!,
      onChange: (v: boolean) => setJobFailed(v),
    },
    {
      id: 'provider-offline',
      title: 'Provider Offline',
      description: 'Notify when an AI provider becomes unavailable',
      enabled: providerOffline!,
      onChange: (v: boolean) => setProviderOffline(v),
    },
    {
      id: 'usage-threshold',
      title: 'Usage Threshold',
      description: 'Notify when you reach 80% of your monthly limit',
      enabled: usageThreshold!,
      onChange: (v: boolean) => setUsageThreshold(v),
    },
    {
      id: 'weekly-summary',
      title: 'Weekly Summary',
      description: 'Receive a weekly email summary of your usage',
      enabled: weeklySummary!,
      onChange: (v: boolean) => setWeeklySummary(v),
    },
  ]

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>Notification Preferences</CardTitle>
          <CardDescription>
            Choose what notifications you want to receive
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {settings.map((setting) => (
            <div
              key={setting.id}
              className="flex items-center justify-between space-x-4"
            >
              <div className="space-y-0.5">
                <Label htmlFor={setting.id}>{setting.title}</Label>
                <p className="text-xs text-muted-foreground">
                  {setting.description}
                </p>
              </div>
              <Switch
                id={setting.id}
                checked={setting.enabled}
                onCheckedChange={setting.onChange}
              />
            </div>
          ))}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Email Notifications</CardTitle>
          <CardDescription>
            Configure email notification settings
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label>Marketing Emails</Label>
              <p className="text-xs text-muted-foreground">
                Receive updates about new features and tips
              </p>
            </div>
            <Switch
              checked={marketingEmails!}
              onCheckedChange={(v) => setMarketingEmails(v)}
            />
          </div>
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label>Security Alerts</Label>
              <p className="text-xs text-muted-foreground">
                Important security notifications (always enabled)
              </p>
            </div>
            <Switch checked disabled />
          </div>
        </CardContent>
      </Card>

      <div className="flex justify-end">
        <Button onClick={handleSave} disabled={updateTenant.isPending || !hasChanges}>
          {updateTenant.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
          Save Preferences
        </Button>
      </div>
    </div>
  )
}
