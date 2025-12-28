'use client'

import { useTranslation } from 'react-i18next'
import { StreamingChat } from '@/components/streaming-chat'

export default function ChatPage() {
  const { t } = useTranslation()

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">{t('chat.title')}</h1>
        <p className="text-muted-foreground">
          {t('chat.hint')}
        </p>
      </div>

      <StreamingChat />
    </div>
  )
}
