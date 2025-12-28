'use client'

import { useTranslation } from 'react-i18next'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { GeneralSettings } from '@/components/settings/general-settings'
import { ApiKeysSettings } from '@/components/settings/api-keys-settings'
import { NotificationsSettings } from '@/components/settings/notifications-settings'
import { PricingSettings } from '@/components/settings/pricing-settings'

export default function SettingsPage() {
  const { t } = useTranslation()

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">{t('settings.title')}</h1>
        <p className="text-muted-foreground">
          {t('settings.tenant.title')}
        </p>
      </div>

      <Tabs defaultValue="general" className="space-y-6">
        <TabsList>
          <TabsTrigger value="general">{t('settings.tenant.title')}</TabsTrigger>
          <TabsTrigger value="api-keys">{t('nav.apiKeys')}</TabsTrigger>
          <TabsTrigger value="notifications">{t('settings.notifications.title')}</TabsTrigger>
          <TabsTrigger value="pricing">{t('nav.pricing')}</TabsTrigger>
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
