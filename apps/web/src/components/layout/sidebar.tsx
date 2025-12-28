'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/utils'
import {
  LayoutDashboard,
  Zap,
  Settings,
  BarChart3,
  Cpu,
  HelpCircle,
  MessageSquare,
} from 'lucide-react'
import { ScrollArea } from '@/components/ui/scroll-area'
import { LanguageSwitcher } from '@/components/language-switcher'

export function Sidebar() {
  const pathname = usePathname()
  const { t } = useTranslation()

  const navigation = [
    { name: t('nav.dashboard'), href: '/', icon: LayoutDashboard },
    { name: t('nav.chat'), href: '/chat', icon: MessageSquare },
    { name: t('nav.jobs'), href: '/jobs', icon: Zap },
    { name: t('nav.providers'), href: '/providers', icon: Cpu },
    { name: t('nav.usage'), href: '/usage', icon: BarChart3 },
    { name: t('nav.settings'), href: '/settings', icon: Settings },
  ]

  const secondaryNavigation = [
    { name: 'Documentation', href: '/docs', icon: HelpCircle },
  ]

  return (
    <div className="flex h-full w-64 flex-col border-r bg-card">
      {/* Logo */}
      <div className="flex h-14 items-center justify-between border-b px-4">
        <Link href="/" className="flex items-center gap-2">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground">
            <Zap className="h-4 w-4" />
          </div>
          <span className="font-semibold">AI Aggregator</span>
        </Link>
        <LanguageSwitcher />
      </div>

      {/* Navigation */}
      <ScrollArea className="flex-1 px-3 py-4">
        <nav className="space-y-1">
          {navigation.map((item) => {
            const isActive = pathname === item.href ||
              (item.href !== '/' && pathname.startsWith(item.href))

            return (
              <Link
                key={item.name}
                href={item.href}
                className={cn(
                  'flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors',
                  isActive
                    ? 'bg-accent text-accent-foreground'
                    : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
                )}
              >
                <item.icon className="h-4 w-4" />
                {item.name}
              </Link>
            )
          })}
        </nav>

        <div className="mt-8">
          <p className="px-3 text-xs font-medium uppercase tracking-wider text-muted-foreground">
            Support
          </p>
          <nav className="mt-2 space-y-1">
            {secondaryNavigation.map((item) => (
              <Link
                key={item.name}
                href={item.href}
                className="flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground"
              >
                <item.icon className="h-4 w-4" />
                {item.name}
              </Link>
            ))}
          </nav>
        </div>
      </ScrollArea>

      {/* Footer */}
      <div className="border-t p-4">
        <div className="flex items-center gap-3">
          <div className="h-8 w-8 rounded-full bg-muted" />
          <div className="flex-1 min-w-0">
            <p className="text-sm font-medium truncate">Default Tenant</p>
            <p className="text-xs text-muted-foreground truncate">
              dev-api-key
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}
