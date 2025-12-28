import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Cpu, Cloud, Server } from 'lucide-react'

const providers = [
  {
    name: 'OpenAI',
    type: 'openai',
    status: 'online',
    latency: '145ms',
    icon: Cloud,
  },
  {
    name: 'Claude',
    type: 'claude',
    status: 'online',
    latency: '198ms',
    icon: Cloud,
  },
  {
    name: 'Stub Provider',
    type: 'local',
    status: 'online',
    latency: '12ms',
    icon: Server,
  },
  {
    name: 'Ollama',
    type: 'ollama',
    status: 'offline',
    latency: '-',
    icon: Cpu,
  },
]

export function ProviderStatus() {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Provider Status</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {providers.map((provider) => (
            <div
              key={provider.name}
              className="flex items-center justify-between rounded-lg border p-3"
            >
              <div className="flex items-center gap-3">
                <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-muted">
                  <provider.icon className="h-5 w-5" />
                </div>
                <div>
                  <p className="text-sm font-medium">{provider.name}</p>
                  <p className="text-xs text-muted-foreground">
                    Latency: {provider.latency}
                  </p>
                </div>
              </div>
              <Badge
                variant={provider.status === 'online' ? 'success' : 'secondary'}
              >
                {provider.status}
              </Badge>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}
