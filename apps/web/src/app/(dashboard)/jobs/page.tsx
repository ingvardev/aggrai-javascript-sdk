import { JobsTable } from '@/components/jobs/jobs-table'
import { CreateJobDialog } from '@/components/jobs/create-job-dialog'
import { Button } from '@/components/ui/button'
import { PlusIcon } from 'lucide-react'

export default function JobsPage() {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Jobs</h1>
          <p className="text-muted-foreground">
            Manage and monitor your AI processing jobs
          </p>
        </div>
        <CreateJobDialog>
          <Button>
            <PlusIcon className="mr-2 h-4 w-4" />
            New Job
          </Button>
        </CreateJobDialog>
      </div>

      <JobsTable />
    </div>
  )
}
