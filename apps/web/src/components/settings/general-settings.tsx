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

export function GeneralSettings() {
  const [tenantName, setTenantName] = useState('Default Tenant')
  const [defaultProvider, setDefaultProvider] = useState('auto')
  const [darkMode, setDarkMode] = useState(true)

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
            <Label htmlFor="tenant-name">Tenant Name</Label>
            <Input
              id="tenant-name"
              value={tenantName}
              onChange={(e) => setTenantName(e.target.value)}
              placeholder="Enter tenant name"
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="default-provider">Default Provider</Label>
            <Select value={defaultProvider} onValueChange={setDefaultProvider}>
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
            <Switch checked={darkMode} onCheckedChange={setDarkMode} />
          </div>
        </CardContent>
      </Card>

      <div className="flex justify-end">
        <Button>Save Changes</Button>
      </div>
    </div>
  )
}
