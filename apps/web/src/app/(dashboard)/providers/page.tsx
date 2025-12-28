import { ProvidersGrid } from '@/components/providers/providers-grid'

export default function ProvidersPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Providers</h1>
        <p className="text-muted-foreground">
          Configure and monitor your AI providers
        </p>
      </div>

      <ProvidersGrid />
    </div>
  )
}
