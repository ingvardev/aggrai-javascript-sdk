'use client'

import { useState } from 'react'
import { useParams, useRouter } from 'next/navigation'
import Link from 'next/link'
import { useTranslation } from 'react-i18next'
import {
  useAPIUser,
  useAPIKeys,
  useCreateAPIKey,
  useRevokeAPIKey
} from '@/lib/hooks'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Skeleton } from '@/components/ui/skeleton'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  ArrowLeft,
  Plus,
  Key,
  Copy,
  Trash2,
  RefreshCw,
  Loader2,
  Eye,
  EyeOff,
  CheckCircle,
  AlertTriangle,
  User,
  Activity,
} from 'lucide-react'
import { formatDate } from '@/lib/utils'
import { toast } from 'sonner'
import { APIKey, APIKeyWithRawKey } from '@/lib/api-users'

export default function APIUserDetailPage() {
  const params = useParams()
  const userId = params.id as string
  const { t } = useTranslation()

  const { data: user, isLoading: isLoadingUser, error: userError } = useAPIUser(userId)
  const { data: keys, isLoading: isLoadingKeys, refetch: refetchKeys } = useAPIKeys(userId)

  const [isCreateKeyDialogOpen, setIsCreateKeyDialogOpen] = useState(false)
  const [newlyCreatedKey, setNewlyCreatedKey] = useState<APIKeyWithRawKey | null>(null)

  if (isLoadingUser) {
    return <UserDetailSkeleton />
  }

  if (userError || !user) {
    return (
      <Card>
        <CardContent className="flex flex-col items-center justify-center py-10">
          <AlertTriangle className="h-12 w-12 text-destructive" />
          <p className="mt-4 text-destructive">{t('users.userNotFound')}</p>
          <Button variant="outline" asChild className="mt-4">
            <Link href="/users">
              <ArrowLeft className="mr-2 h-4 w-4" />
              {t('users.backToList')}
            </Link>
          </Button>
        </CardContent>
      </Card>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <Button variant="ghost" size="icon" asChild>
          <Link href="/users">
            <ArrowLeft className="h-4 w-4" />
          </Link>
        </Button>
        <div className="flex-1">
          <div className="flex items-center gap-2">
            <h1 className="text-3xl font-bold tracking-tight">{user.name}</h1>
            <Badge variant={user.active ? 'success' : 'secondary'}>
              {user.active ? t('users.active') : t('users.inactive')}
            </Badge>
          </div>
          <p className="text-muted-foreground">{user.description || t('users.noDescription')}</p>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">{t('users.apiKeysCount')}</CardTitle>
            <Key className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{keys?.length ?? 0}</div>
            <p className="text-xs text-muted-foreground">
              {keys?.filter(k => k.active).length ?? 0} {t('users.active')}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">{t('users.createdAt')}</CardTitle>
            <User className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{formatDate(user.created_at)}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">{t('users.lastUpdated')}</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{formatDate(user.updated_at)}</div>
          </CardContent>
        </Card>
      </div>

      {/* Tabs */}
      <Tabs defaultValue="keys" className="space-y-4">
        <TabsList>
          <TabsTrigger value="keys" id="keys">{t('users.apiKeys')}</TabsTrigger>
          <TabsTrigger value="activity">{t('users.activity')}</TabsTrigger>
        </TabsList>

        <TabsContent value="keys" className="space-y-4">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between">
              <div>
                <CardTitle>{t('users.apiKeys')}</CardTitle>
                <CardDescription>{t('users.apiKeysDescription')}</CardDescription>
              </div>
              <div className="flex items-center gap-2">
                <Button variant="outline" size="icon" onClick={() => refetchKeys()}>
                  <RefreshCw className="h-4 w-4" />
                </Button>
                <CreateKeyDialog
                  userId={userId}
                  open={isCreateKeyDialogOpen}
                  onOpenChange={setIsCreateKeyDialogOpen}
                  onKeyCreated={(key) => setNewlyCreatedKey(key)}
                />
              </div>
            </CardHeader>
            <CardContent>
              {isLoadingKeys ? (
                <div className="space-y-4">
                  {[...Array(3)].map((_, i) => (
                    <Skeleton key={i} className="h-16 w-full" />
                  ))}
                </div>
              ) : keys && keys.length > 0 ? (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>{t('users.keyName')}</TableHead>
                      <TableHead>{t('users.keyPrefix')}</TableHead>
                      <TableHead>{t('users.scopes')}</TableHead>
                      <TableHead>{t('common.status')}</TableHead>
                      <TableHead>{t('users.lastUsed')}</TableHead>
                      <TableHead>{t('users.usageCount')}</TableHead>
                      <TableHead className="text-right">{t('common.actions')}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {keys.map((key) => (
                      <APIKeyRow key={key.id} apiKey={key} userId={userId} />
                    ))}
                  </TableBody>
                </Table>
              ) : (
                <div className="flex flex-col items-center justify-center py-10 text-center">
                  <Key className="h-12 w-12 text-muted-foreground" />
                  <h3 className="mt-4 text-lg font-semibold">{t('users.noKeys')}</h3>
                  <p className="mt-2 text-sm text-muted-foreground">
                    {t('users.noKeysDescription')}
                  </p>
                  <Button className="mt-4" onClick={() => setIsCreateKeyDialogOpen(true)}>
                    <Plus className="mr-2 h-4 w-4" />
                    {t('users.createKey')}
                  </Button>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="activity">
          <Card>
            <CardHeader>
              <CardTitle>{t('users.activity')}</CardTitle>
              <CardDescription>{t('users.activityDescription')}</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="flex flex-col items-center justify-center py-10 text-center">
                <Activity className="h-12 w-12 text-muted-foreground" />
                <h3 className="mt-4 text-lg font-semibold">{t('users.noActivity')}</h3>
                <p className="mt-2 text-sm text-muted-foreground">
                  {t('users.noActivityDescription')}
                </p>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* New Key Dialog */}
      <NewKeyCreatedDialog
        apiKey={newlyCreatedKey}
        open={!!newlyCreatedKey}
        onOpenChange={(open) => !open && setNewlyCreatedKey(null)}
      />
    </div>
  )
}

function APIKeyRow({ apiKey, userId }: { apiKey: APIKey; userId: string }) {
  const { t } = useTranslation()
  const revokeKey = useRevokeAPIKey()
  const [showPrefix, setShowPrefix] = useState(false)

  const handleRevoke = async () => {
    try {
      await revokeKey.mutateAsync({ keyId: apiKey.id, userId })
      toast.success(t('users.keyRevoked'))
    } catch (error) {
      toast.error(t('users.revokeError'))
    }
  }

  return (
    <TableRow>
      <TableCell className="font-medium">{apiKey.name}</TableCell>
      <TableCell>
        <div className="flex items-center gap-2">
          <code className="rounded bg-muted px-2 py-1 text-sm">
            {showPrefix ? apiKey.key_prefix : '••••••••'}
          </code>
          <Button
            variant="ghost"
            size="icon"
            className="h-6 w-6"
            onClick={() => setShowPrefix(!showPrefix)}
          >
            {showPrefix ? <EyeOff className="h-3 w-3" /> : <Eye className="h-3 w-3" />}
          </Button>
        </div>
      </TableCell>
      <TableCell>
        <div className="flex flex-wrap gap-1">
          {apiKey.scopes?.map((scope) => (
            <Badge key={scope} variant="outline" className="text-xs">
              {scope}
            </Badge>
          )) ?? (
            <Badge variant="outline" className="text-xs">
              {t('users.allScopes')}
            </Badge>
          )}
        </div>
      </TableCell>
      <TableCell>
        <Badge variant={apiKey.active ? 'success' : 'destructive'}>
          {apiKey.active ? t('users.active') : t('users.revoked')}
        </Badge>
      </TableCell>
      <TableCell>
        {apiKey.last_used_at ? formatDate(apiKey.last_used_at) : t('users.never')}
      </TableCell>
      <TableCell>{apiKey.usage_count}</TableCell>
      <TableCell className="text-right">
        {apiKey.active && (
          <AlertDialog>
            <AlertDialogTrigger asChild>
              <Button variant="ghost" size="icon" className="text-destructive">
                <Trash2 className="h-4 w-4" />
              </Button>
            </AlertDialogTrigger>
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>{t('users.revokeKeyTitle')}</AlertDialogTitle>
                <AlertDialogDescription>
                  {t('users.revokeKeyDescription')}
                </AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel>{t('common.cancel')}</AlertDialogCancel>
                <AlertDialogAction
                  onClick={handleRevoke}
                  className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                >
                  {revokeKey.isPending && (
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  )}
                  {t('users.revokeKey')}
                </AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>
        )}
      </TableCell>
    </TableRow>
  )
}

function CreateKeyDialog({
  userId,
  open,
  onOpenChange,
  onKeyCreated,
}: {
  userId: string
  open: boolean
  onOpenChange: (open: boolean) => void
  onKeyCreated: (key: APIKeyWithRawKey) => void
}) {
  const { t } = useTranslation()
  const createKey = useCreateAPIKey()
  const [name, setName] = useState('')
  const [scopes, setScopes] = useState('read,write')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    try {
      const scopesList = scopes.split(',').map(s => s.trim()).filter(Boolean)
      const key = await createKey.mutateAsync({ userId, name, scopes: scopesList })
      onKeyCreated(key)
      onOpenChange(false)
      setName('')
      setScopes('read,write')
    } catch (error) {
      toast.error(t('users.createKeyError'))
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogTrigger asChild>
        <Button>
          <Plus className="mr-2 h-4 w-4" />
          {t('users.createKey')}
        </Button>
      </DialogTrigger>
      <DialogContent>
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>{t('users.createKey')}</DialogTitle>
            <DialogDescription>
              {t('users.createKeyDescription')}
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="keyName">{t('users.keyName')}</Label>
              <Input
                id="keyName"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder={t('users.keyNamePlaceholder')}
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="scopes">{t('users.scopes')}</Label>
              <Input
                id="scopes"
                value={scopes}
                onChange={(e) => setScopes(e.target.value)}
                placeholder="read,write,admin"
              />
              <p className="text-xs text-muted-foreground">
                {t('users.scopesHelp')}
              </p>
            </div>
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              {t('common.cancel')}
            </Button>
            <Button type="submit" disabled={createKey.isPending}>
              {createKey.isPending && (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              )}
              {t('common.create')}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

function NewKeyCreatedDialog({
  apiKey,
  open,
  onOpenChange,
}: {
  apiKey: APIKeyWithRawKey | null
  open: boolean
  onOpenChange: (open: boolean) => void
}) {
  const { t } = useTranslation()
  const [copied, setCopied] = useState(false)

  const handleCopy = async () => {
    if (apiKey?.key) {
      await navigator.clipboard.writeText(apiKey.key)
      setCopied(true)
      toast.success(t('users.keyCopied'))
      setTimeout(() => setCopied(false), 2000)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <CheckCircle className="h-5 w-5 text-green-500" />
            {t('users.keyCreated')}
          </DialogTitle>
          <DialogDescription>
            {t('users.keyCreatedDescription')}
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4 py-4">
          <div className="rounded-lg border bg-muted p-4">
            <div className="flex items-center justify-between gap-2">
              <code className="flex-1 break-all text-sm">{apiKey?.key}</code>
              <Button variant="outline" size="icon" onClick={handleCopy}>
                {copied ? (
                  <CheckCircle className="h-4 w-4 text-green-500" />
                ) : (
                  <Copy className="h-4 w-4" />
                )}
              </Button>
            </div>
          </div>
          <div className="flex items-center gap-2 rounded-lg border border-yellow-500/50 bg-yellow-500/10 p-3">
            <AlertTriangle className="h-5 w-5 text-yellow-500" />
            <p className="text-sm text-yellow-600 dark:text-yellow-400">
              {t('users.keyWarning')}
            </p>
          </div>
        </div>
        <DialogFooter>
          <Button onClick={() => onOpenChange(false)}>
            {t('common.close')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

function UserDetailSkeleton() {
  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Skeleton className="h-10 w-10" />
        <div className="flex-1 space-y-2">
          <Skeleton className="h-8 w-48" />
          <Skeleton className="h-4 w-32" />
        </div>
      </div>
      <div className="grid gap-4 md:grid-cols-3">
        {[...Array(3)].map((_, i) => (
          <Skeleton key={i} className="h-28" />
        ))}
      </div>
      <Skeleton className="h-96" />
    </div>
  )
}
