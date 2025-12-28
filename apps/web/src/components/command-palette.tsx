'use client'

import * as React from 'react'
import { useRouter } from 'next/navigation'
import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
  CommandShortcut,
} from '@/components/ui/command'
import {
  LayoutDashboard,
  Briefcase,
  Server,
  Settings,
  Plus,
  Search,
  Moon,
  Sun,
  RefreshCw,
  Copy,
  ExternalLink,
} from 'lucide-react'
import { useRecentJobs } from '@/lib/hooks'
import { toast } from 'sonner'

export function CommandPalette() {
  const [open, setOpen] = React.useState(false)
  const router = useRouter()
  const { data: recentJobs } = useRecentJobs()

  React.useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if (e.key === 'k' && (e.metaKey || e.ctrlKey)) {
        e.preventDefault()
        setOpen((open) => !open)
      }
    }

    document.addEventListener('keydown', down)
    return () => document.removeEventListener('keydown', down)
  }, [])

  const runCommand = React.useCallback((command: () => void) => {
    setOpen(false)
    command()
  }, [])

  const navigateTo = (path: string) => {
    runCommand(() => router.push(path))
  }

  const toggleTheme = () => {
    runCommand(() => {
      document.documentElement.classList.toggle('dark')
      toast.success('Theme toggled')
    })
  }

  const copyApiKey = () => {
    runCommand(() => {
      navigator.clipboard.writeText('dev-api-key-12345')
      toast.success('API key copied to clipboard')
    })
  }

  const openPlayground = () => {
    runCommand(() => {
      window.open('http://localhost:8080/playground', '_blank')
      toast.info('Opening GraphQL Playground')
    })
  }

  const refreshPage = () => {
    runCommand(() => {
      window.location.reload()
      toast.info('Refreshing...')
    })
  }

  return (
    <CommandDialog open={open} onOpenChange={setOpen}>
      <CommandInput placeholder="Type a command or search..." />
      <CommandList>
        <CommandEmpty>No results found.</CommandEmpty>

        <CommandGroup heading="Navigation">
          <CommandItem onSelect={() => navigateTo('/')}>
            <LayoutDashboard className="mr-2 h-4 w-4" />
            <span>Dashboard</span>
            <CommandShortcut>⌘D</CommandShortcut>
          </CommandItem>
          <CommandItem onSelect={() => navigateTo('/jobs')}>
            <Briefcase className="mr-2 h-4 w-4" />
            <span>Jobs</span>
            <CommandShortcut>⌘J</CommandShortcut>
          </CommandItem>
          <CommandItem onSelect={() => navigateTo('/providers')}>
            <Server className="mr-2 h-4 w-4" />
            <span>Providers</span>
            <CommandShortcut>⌘P</CommandShortcut>
          </CommandItem>
          <CommandItem onSelect={() => navigateTo('/settings')}>
            <Settings className="mr-2 h-4 w-4" />
            <span>Settings</span>
            <CommandShortcut>⌘,</CommandShortcut>
          </CommandItem>
        </CommandGroup>

        <CommandSeparator />

        <CommandGroup heading="Actions">
          <CommandItem onSelect={() => navigateTo('/jobs?new=true')}>
            <Plus className="mr-2 h-4 w-4" />
            <span>Create New Job</span>
            <CommandShortcut>⌘N</CommandShortcut>
          </CommandItem>
          <CommandItem onSelect={copyApiKey}>
            <Copy className="mr-2 h-4 w-4" />
            <span>Copy API Key</span>
          </CommandItem>
          <CommandItem onSelect={openPlayground}>
            <ExternalLink className="mr-2 h-4 w-4" />
            <span>Open GraphQL Playground</span>
          </CommandItem>
          <CommandItem onSelect={refreshPage}>
            <RefreshCw className="mr-2 h-4 w-4" />
            <span>Refresh</span>
            <CommandShortcut>⌘R</CommandShortcut>
          </CommandItem>
        </CommandGroup>

        <CommandSeparator />

        <CommandGroup heading="Preferences">
          <CommandItem onSelect={toggleTheme}>
            <Sun className="mr-2 h-4 w-4 dark:hidden" />
            <Moon className="mr-2 h-4 w-4 hidden dark:block" />
            <span>Toggle Theme</span>
            <CommandShortcut>⌘T</CommandShortcut>
          </CommandItem>
        </CommandGroup>

        {recentJobs && recentJobs.length > 0 && (
          <>
            <CommandSeparator />
            <CommandGroup heading="Recent Jobs">
              {recentJobs.slice(0, 5).map((job) => (
                <CommandItem
                  key={job.id}
                  onSelect={() => navigateTo(`/jobs/${job.id}`)}
                >
                  <Search className="mr-2 h-4 w-4" />
                  <span className="truncate flex-1">
                    {job.input.slice(0, 50)}{job.input.length > 50 ? '...' : ''}
                  </span>
                  <span className={`text-xs ml-2 ${
                    job.status === 'COMPLETED' ? 'text-green-500' :
                    job.status === 'FAILED' ? 'text-red-500' :
                    job.status === 'PROCESSING' ? 'text-blue-500' :
                    'text-yellow-500'
                  }`}>
                    {job.status}
                  </span>
                </CommandItem>
              ))}
            </CommandGroup>
          </>
        )}
      </CommandList>
    </CommandDialog>
  )
}

// Keyboard shortcut hint component
export function CommandPaletteHint() {
  return (
    <button
      onClick={() => {
        const event = new KeyboardEvent('keydown', {
          key: 'k',
          metaKey: true,
          bubbles: true,
        })
        document.dispatchEvent(event)
      }}
      className="hidden md:flex items-center gap-2 px-3 py-1.5 text-sm text-muted-foreground bg-muted/50 rounded-md border hover:bg-muted transition-colors"
    >
      <Search className="h-3.5 w-3.5" />
      <span>Search...</span>
      <kbd className="pointer-events-none ml-2 inline-flex h-5 select-none items-center gap-1 rounded border bg-muted px-1.5 font-mono text-[10px] font-medium">
        <span className="text-xs">⌘</span>K
      </kbd>
    </button>
  )
}
