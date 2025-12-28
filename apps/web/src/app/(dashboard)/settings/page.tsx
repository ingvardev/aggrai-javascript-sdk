import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { GeneralSettings } from '@/components/settings/general-settings'
import { ApiKeysSettings } from '@/components/settings/api-keys-settings'
import { NotificationsSettings } from '@/components/settings/notifications-settings'
import { PricingSettings } from '@/components/settings/pricing-settings'

export default function SettingsPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Settings</h1>
        <p className="text-muted-foreground">
          Manage your account and application preferences
        </p>
      </div>

      <Tabs defaultValue="general" className="space-y-6">
        <TabsList>
          <TabsTrigger value="general">General</TabsTrigger>
          <TabsTrigger value="api-keys">API Keys</TabsTrigger>
          <TabsTrigger value="notifications">Notifications</TabsTrigger>
          <TabsTrigger value="pricing">Pricing</TabsTrigger>
        </TabsList>

        <TabsContent value="general" className="space-y-6">
          <GeneralSettings />
        </TabsContent>

        <TabsContent value="api-keys" className="space-y-6">
          <ApiKeysSettings />
        </TabsContent>

        <TabsContent value="notifications" className="space-y-6">
          <NotificationsSettings />
        </TabsContent>

        <TabsContent value="pricing" className="space-y-6">
          <PricingSettings />
        </TabsContent>
      </Tabs>
    </div>
  )
}
