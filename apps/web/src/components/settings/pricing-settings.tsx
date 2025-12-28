'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Skeleton } from '@/components/ui/skeleton'
import { Badge } from '@/components/ui/badge'
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
import { usePricingList, useUpdatePricing, useCreatePricing, useDeletePricing } from '@/lib/hooks'
import { getProviderDisplayName } from '@/lib/utils'
import { toast } from 'sonner'
import { Loader2, Plus, Pencil, Trash2, DollarSign } from 'lucide-react'

interface PricingFormData {
  provider: string
  model: string
  inputPricePerMillion: number
  outputPricePerMillion: number
  imagePrice: number | null
  isDefault: boolean
}

export function PricingSettings() {
  const { data: pricingList, isLoading } = usePricingList()
  const updatePricing = useUpdatePricing()
  const createPricing = useCreatePricing()
  const deletePricing = useDeletePricing()

  const [isDialogOpen, setIsDialogOpen] = useState(false)
  const [editingPricing, setEditingPricing] = useState<{ id: string } & PricingFormData | null>(null)
  const [formData, setFormData] = useState<PricingFormData>({
    provider: '',
    model: '',
    inputPricePerMillion: 0,
    outputPricePerMillion: 0,
    imagePrice: null,
    isDefault: false,
  })

  const handleEdit = (pricing: typeof pricingList[0]) => {
    setEditingPricing({
      id: pricing.id,
      provider: pricing.provider,
      model: pricing.model,
      inputPricePerMillion: pricing.inputPricePerMillion,
      outputPricePerMillion: pricing.outputPricePerMillion,
      imagePrice: pricing.imagePrice,
      isDefault: pricing.isDefault,
    })
    setFormData({
      provider: pricing.provider,
      model: pricing.model,
      inputPricePerMillion: pricing.inputPricePerMillion,
      outputPricePerMillion: pricing.outputPricePerMillion,
      imagePrice: pricing.imagePrice,
      isDefault: pricing.isDefault,
    })
    setIsDialogOpen(true)
  }

  const handleCreate = () => {
    setEditingPricing(null)
    setFormData({
      provider: '',
      model: '',
      inputPricePerMillion: 0,
      outputPricePerMillion: 0,
      imagePrice: null,
      isDefault: false,
    })
    setIsDialogOpen(true)
  }

  const handleSave = async () => {
    try {
      if (editingPricing) {
        await updatePricing.mutateAsync({
          id: editingPricing.id,
          inputPricePerMillion: formData.inputPricePerMillion,
          outputPricePerMillion: formData.outputPricePerMillion,
          imagePrice: formData.imagePrice,
          isDefault: formData.isDefault,
        })
        toast.success('Pricing updated successfully')
      } else {
        await createPricing.mutateAsync(formData)
        toast.success('Pricing created successfully')
      }
      setIsDialogOpen(false)
    } catch (error) {
      toast.error('Failed to save pricing')
    }
  }

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this pricing configuration?')) {
      return
    }
    try {
      await deletePricing.mutateAsync(id)
      toast.success('Pricing deleted successfully')
    } catch (error) {
      toast.error('Failed to delete pricing')
    }
  }

  // Group pricing by provider
  const groupedPricing = pricingList?.reduce((acc, pricing) => {
    if (!acc[pricing.provider]) {
      acc[pricing.provider] = []
    }
    acc[pricing.provider].push(pricing)
    return acc
  }, {} as Record<string, typeof pricingList>)

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <Skeleton className="h-6 w-48" />
          <Skeleton className="h-4 w-64" />
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {[1, 2, 3].map((i) => (
              <Skeleton key={i} className="h-12 w-full" />
            ))}
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="flex items-center gap-2">
              <DollarSign className="h-5 w-5" />
              Provider Pricing
            </CardTitle>
            <CardDescription>
              Configure pricing per million tokens for each AI provider and model
            </CardDescription>
          </div>
          <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
            <DialogTrigger asChild>
              <Button onClick={handleCreate}>
                <Plus className="mr-2 h-4 w-4" />
                Add Pricing
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>
                  {editingPricing ? 'Edit Pricing' : 'Add Pricing'}
                </DialogTitle>
                <DialogDescription>
                  {editingPricing
                    ? 'Update the pricing configuration for this model'
                    : 'Add a new pricing configuration for a provider model'}
                </DialogDescription>
              </DialogHeader>
              <div className="grid gap-4 py-4">
                {!editingPricing && (
                  <>
                    <div className="grid gap-2">
                      <Label htmlFor="provider">Provider</Label>
                      <Input
                        id="provider"
                        placeholder="e.g., openai, claude"
                        value={formData.provider}
                        onChange={(e) =>
                          setFormData({ ...formData, provider: e.target.value })
                        }
                      />
                    </div>
                    <div className="grid gap-2">
                      <Label htmlFor="model">Model</Label>
                      <Input
                        id="model"
                        placeholder="e.g., gpt-4o-mini"
                        value={formData.model}
                        onChange={(e) =>
                          setFormData({ ...formData, model: e.target.value })
                        }
                      />
                    </div>
                  </>
                )}
                <div className="grid gap-2">
                  <Label htmlFor="inputPrice">Input Price (per 1M tokens)</Label>
                  <Input
                    id="inputPrice"
                    type="number"
                    step="0.01"
                    min="0"
                    value={formData.inputPricePerMillion}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        inputPricePerMillion: parseFloat(e.target.value) || 0,
                      })
                    }
                  />
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="outputPrice">Output Price (per 1M tokens)</Label>
                  <Input
                    id="outputPrice"
                    type="number"
                    step="0.01"
                    min="0"
                    value={formData.outputPricePerMillion}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        outputPricePerMillion: parseFloat(e.target.value) || 0,
                      })
                    }
                  />
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="imagePrice">Image Price (optional)</Label>
                  <Input
                    id="imagePrice"
                    type="number"
                    step="0.01"
                    min="0"
                    placeholder="Leave empty if not an image model"
                    value={formData.imagePrice ?? ''}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        imagePrice: e.target.value ? parseFloat(e.target.value) : null,
                      })
                    }
                  />
                </div>
                <div className="flex items-center gap-2">
                  <input
                    type="checkbox"
                    id="isDefault"
                    checked={formData.isDefault}
                    onChange={(e) =>
                      setFormData({ ...formData, isDefault: e.target.checked })
                    }
                    className="h-4 w-4 rounded border-gray-300"
                  />
                  <Label htmlFor="isDefault">Set as default for this provider</Label>
                </div>
              </div>
              <DialogFooter>
                <Button variant="outline" onClick={() => setIsDialogOpen(false)}>
                  Cancel
                </Button>
                <Button
                  onClick={handleSave}
                  disabled={updatePricing.isPending || createPricing.isPending}
                >
                  {(updatePricing.isPending || createPricing.isPending) && (
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  )}
                  Save
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </div>
      </CardHeader>
      <CardContent>
        {groupedPricing && Object.keys(groupedPricing).length > 0 ? (
          <div className="space-y-6">
            {Object.entries(groupedPricing).map(([provider, models]) => (
              <div key={provider}>
                <h3 className="mb-2 font-semibold">{getProviderDisplayName(provider)}</h3>
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Model</TableHead>
                      <TableHead className="text-right">Input $/1M</TableHead>
                      <TableHead className="text-right">Output $/1M</TableHead>
                      <TableHead className="text-right">Image $</TableHead>
                      <TableHead className="text-center">Default</TableHead>
                      <TableHead className="text-right">Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {models.map((pricing) => (
                      <TableRow key={pricing.id}>
                        <TableCell className="font-mono text-sm">
                          {pricing.model}
                        </TableCell>
                        <TableCell className="text-right">
                          ${pricing.inputPricePerMillion.toFixed(2)}
                        </TableCell>
                        <TableCell className="text-right">
                          ${pricing.outputPricePerMillion.toFixed(2)}
                        </TableCell>
                        <TableCell className="text-right">
                          {pricing.imagePrice != null
                            ? `$${pricing.imagePrice.toFixed(2)}`
                            : '-'}
                        </TableCell>
                        <TableCell className="text-center">
                          {pricing.isDefault && (
                            <Badge variant="secondary">Default</Badge>
                          )}
                        </TableCell>
                        <TableCell className="text-right">
                          <div className="flex justify-end gap-2">
                            <Button
                              variant="ghost"
                              size="icon"
                              onClick={() => handleEdit(pricing)}
                            >
                              <Pencil className="h-4 w-4" />
                            </Button>
                            <Button
                              variant="ghost"
                              size="icon"
                              onClick={() => handleDelete(pricing.id)}
                            >
                              <Trash2 className="h-4 w-4 text-destructive" />
                            </Button>
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            ))}
          </div>
        ) : (
          <p className="text-sm text-muted-foreground">
            No pricing configurations found. Add your first pricing configuration above.
          </p>
        )}
      </CardContent>
    </Card>
  )
}
