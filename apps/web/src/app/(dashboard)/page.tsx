import { StatsCards } from '@/components/dashboard/stats-cards'
import { RecentJobs } from '@/components/dashboard/recent-jobs'
import { ProviderStatus } from '@/components/dashboard/provider-status'

export default function DashboardPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Dashboard</h1>
        <p className="text-muted-foreground">
          Overview of your AI usage and recent activity
        </p>
      </div>

      <StatsCards />

      <div className="grid gap-6 lg:grid-cols-2">
        <RecentJobs />
        <ProviderStatus />
      </div>
    </div>
  )
}
