'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Loader2 } from 'lucide-react'
import { useCreateJob } from '@/lib/hooks'
import { toast } from 'sonner'

interface CreateJobDialogProps {
  children: React.ReactNode
}

export function CreateJobDialog({ children }: CreateJobDialogProps) {
  const [open, setOpen] = useState(false)
  const [input, setInput] = useState('')
  const [type, setType] = useState<'TEXT' | 'IMAGE'>('TEXT')
  const createJob = useCreateJob()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    try {
      const job = await createJob.mutateAsync({ type, input })
      setOpen(false)
      setInput('')
      setType('TEXT')
      toast.success('Job created successfully', {
        description: `Job ID: ${job.id.slice(0, 8)}...`,
        action: {
          label: 'View',
          onClick: () => window.location.href = `/jobs/${job.id}`,
        },
      })
    } catch (error) {
      console.error('Failed to create job:', error)
      toast.error('Failed to create job', {
        description: 'Check if the API server is running.',
      })
    }
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="sm:max-w-[500px]">
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>Create New Job</DialogTitle>
            <DialogDescription>
              Submit a new AI processing request. Choose a type and enter your
              prompt.
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label>Job Type</Label>
              <div className="flex gap-2">
                <Button
                  type="button"
                  variant={type === 'TEXT' ? 'default' : 'outline'}
                  className="flex-1"
                  onClick={() => setType('TEXT')}
                >
                  Text
                </Button>
                <Button
                  type="button"
                  variant={type === 'IMAGE' ? 'default' : 'outline'}
                  className="flex-1"
                  onClick={() => setType('IMAGE')}
                >
                  Image
                </Button>
              </div>
            </div>
            <div className="grid gap-2">
              <Label htmlFor="input">Prompt</Label>
              <Textarea
                id="input"
                placeholder={
                  type === 'TEXT'
                    ? 'Enter your text prompt...'
                    : 'Describe the image you want to generate...'
                }
                value={input}
                onChange={(e) => setInput(e.target.value)}
                className="min-h-[120px]"
                required
              />
            </div>
            {createJob.error && (
              <p className="text-sm text-destructive">
                Failed to create job. Check if the API is running.
              </p>
            )}
          </div>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => setOpen(false)}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={createJob.isPending || !input.trim()}>
              {createJob.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Create Job
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
