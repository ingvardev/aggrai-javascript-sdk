import { JobDetails } from '@/components/jobs/job-details'

interface JobPageProps {
  params: {
    id: string
  }
}

export default function JobPage({ params }: JobPageProps) {
  return (
    <div className="space-y-6">
      <JobDetails jobId={params.id} />
    </div>
  )
}
