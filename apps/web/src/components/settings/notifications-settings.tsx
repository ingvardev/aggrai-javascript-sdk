'use client'

import { useState } from 'react'
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

interface NotificationSetting {
  id: string
  title: string
  description: string
  enabled: boolean
}

const defaultSettings: NotificationSetting[] = [
  {
    id: 'job-completed',
    title: 'Job Completed',
    description: 'Notify when a job finishes processing',
    enabled: true,
  },
  {
    id: 'job-failed',
    title: 'Job Failed',
    description: 'Notify when a job fails to process',
    enabled: true,
  },
  {
    id: 'provider-offline',
    title: 'Provider Offline',
    description: 'Notify when an AI provider becomes unavailable',
    enabled: true,
  },
  {
    id: 'usage-threshold',
    title: 'Usage Threshold',
    description: 'Notify when you reach 80% of your monthly limit',
    enabled: false,
  },
  {
    id: 'weekly-summary',
    title: 'Weekly Summary',
    description: 'Receive a weekly email summary of your usage',
    enabled: false,
  },
]

export function NotificationsSettings() {
  const [settings, setSettings] = useState(defaultSettings)

  const toggleSetting = (id: string) => {
    setSettings((prev) =>
      prev.map((setting) =>
        setting.id === id ? { ...setting, enabled: !setting.enabled } : setting
      )
    )
  }

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
                onCheckedChange={() => toggleSetting(setting.id)}
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
            <Switch />
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
        <Button>Save Preferences</Button>
      </div>
    </div>
  )
}
