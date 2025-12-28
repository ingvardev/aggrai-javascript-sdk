import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Zap, CheckCircle, Clock, DollarSign } from 'lucide-react'

const stats = [
  {
    title: 'Total Jobs',
    value: '1,234',
    description: '+12% from last month',
    icon: Zap,
  },
  {
    title: 'Completed',
    value: '1,180',
    description: '95.6% success rate',
    icon: CheckCircle,
  },
  {
    title: 'Avg. Processing',
    value: '2.4s',
    description: '-0.3s from last week',
    icon: Clock,
  },
  {
    title: 'Total Cost',
    value: '$45.32',
    description: 'This month',
    icon: DollarSign,
  },
]

export function StatsCards() {
  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      {stats.map((stat) => (
        <Card key={stat.title}>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              {stat.title}
            </CardTitle>
            <stat.icon className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stat.value}</div>
            <p className="text-xs text-muted-foreground">
              {stat.description}
            </p>
          </CardContent>
        </Card>
      ))}
    </div>
  )
}
