'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
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

const navigation = [
  { name: 'Dashboard', href: '/', icon: LayoutDashboard },
  { name: 'Chat', href: '/chat', icon: MessageSquare },
  { name: 'Jobs', href: '/jobs', icon: Zap },
  { name: 'Providers', href: '/providers', icon: Cpu },
  { name: 'Usage', href: '/usage', icon: BarChart3 },
  { name: 'Settings', href: '/settings', icon: Settings },
]

const secondaryNavigation = [
  { name: 'Documentation', href: '/docs', icon: HelpCircle },
]

export function Sidebar() {
  const pathname = usePathname()

  return (
    <div className="flex h-full w-64 flex-col border-r bg-card">
      {/* Logo */}
      <div className="flex h-14 items-center border-b px-4">
        <Link href="/" className="flex items-center gap-2">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground">
            <Zap className="h-4 w-4" />
          </div>
          <span className="font-semibold">AI Aggregator</span>
        </Link>
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
