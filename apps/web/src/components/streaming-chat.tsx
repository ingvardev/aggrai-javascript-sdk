'use client'

import { useState, useRef, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Textarea } from '@/components/ui/textarea'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { useStreaming, useProviders } from '@/lib/hooks'
import { useProviderModels } from '@/lib/hooks/use-provider-models'
import { formatCurrency, getProviderDisplayName } from '@/lib/utils'
import { Loader2, Send, Square, RotateCcw, Sparkles } from 'lucide-react'

export function StreamingChat() {
  const [prompt, setPrompt] = useState('')
  const [selectedProvider, setSelectedProvider] = useState<string>('')
  const [selectedModel, setSelectedModel] = useState<string>('')
  const { data: providers } = useProviders()
  const { data: models, isLoading: isLoadingModels } = useProviderModels(selectedProvider || null)
  const contentRef = useRef<HTMLDivElement>(null)

  const {
    content,
    isStreaming,
    error,
    stats,
    startStream,
    stopStream,
    reset,
  } = useStreaming()

  // Auto-scroll as content streams in
  useEffect(() => {
    if (contentRef.current) {
      contentRef.current.scrollTop = contentRef.current.scrollHeight
    }
  }, [content])

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!prompt.trim() || isStreaming) return

    startStream({
      prompt: prompt.trim(),
      provider: selectedProvider || undefined,
      model: selectedModel || undefined,
    })
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSubmit(e)
    }
  }

  const handleReset = () => {
    reset()
    setPrompt('')
  }

  // Filter to text-capable providers (exclude image-only)
  const textProviders = providers?.filter(
    (p) => p.type !== 'LOCAL' || p.name !== 'dall-e'
  )

  // Reset model when provider changes
  useEffect(() => {
    setSelectedModel('')
  }, [selectedProvider])

  return (
    <Card className="flex h-[600px] flex-col">
      <CardHeader className="flex-shrink-0 pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2">
            <Sparkles className="h-5 w-5" />
            Streaming Chat
          </CardTitle>
          <div className="flex items-center gap-2">
            <Select
              value={selectedProvider || 'auto'}
              onValueChange={(value) => setSelectedProvider(value === 'auto' ? '' : value)}
            >
              <SelectTrigger className="w-[180px]">
                <SelectValue placeholder="Auto-select provider" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="auto">Auto-select</SelectItem>
                {textProviders?.map((provider) => (
                  <SelectItem key={provider.name} value={provider.name}>
                    {getProviderDisplayName(provider.name)}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {selectedProvider && (
              <Select
                value={selectedModel || 'default'}
                onValueChange={(value) => setSelectedModel(value === 'default' ? '' : value)}
                disabled={isLoadingModels}
              >
                <SelectTrigger className="w-[200px]">
                  <SelectValue placeholder={isLoadingModels ? 'Loading...' : 'Default model'} />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="default">Default model</SelectItem>
                  {models?.map((model) => (
                    <SelectItem key={model.id} value={model.id}>
                      {model.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
          </div>
        </div>
      </CardHeader>

      <CardContent className="flex flex-1 flex-col gap-4 overflow-hidden">
        {/* Response area */}
        <div
          ref={contentRef}
          className="flex-1 overflow-auto rounded-lg border bg-muted/30 p-4"
        >
          {content ? (
            <div className="whitespace-pre-wrap font-mono text-sm">
              {content}
              {isStreaming && (
                <span className="ml-1 inline-block h-4 w-2 animate-pulse bg-primary" />
              )}
            </div>
          ) : error ? (
            <div className="text-destructive">{error}</div>
          ) : (
            <div className="text-muted-foreground">
              Enter a prompt below and watch the response stream in real-time...
            </div>
          )}
        </div>

        {/* Stats bar */}
        {stats && (
          <div className="flex items-center gap-4 text-sm text-muted-foreground">
            <Badge variant="outline">
              Tokens In: {stats.tokensIn.toLocaleString()}
            </Badge>
            <Badge variant="outline">
              Tokens Out: {stats.tokensOut.toLocaleString()}
            </Badge>
            <Badge variant="outline">
              Cost: {formatCurrency(stats.cost)}
            </Badge>
          </div>
        )}

        {/* Input area */}
        <form onSubmit={handleSubmit} className="flex gap-2">
          <Textarea
            value={prompt}
            onChange={(e) => setPrompt(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Enter your prompt... (Enter to send, Shift+Enter for new line)"
            className="min-h-[80px] flex-1 resize-none"
            disabled={isStreaming}
          />
          <div className="flex flex-col gap-2">
            {isStreaming ? (
              <Button
                type="button"
                variant="destructive"
                size="icon"
                onClick={stopStream}
              >
                <Square className="h-4 w-4" />
              </Button>
            ) : (
              <Button
                type="submit"
                size="icon"
                disabled={!prompt.trim()}
              >
                <Send className="h-4 w-4" />
              </Button>
            )}
            <Button
              type="button"
              variant="outline"
              size="icon"
              onClick={handleReset}
              disabled={isStreaming && !content}
            >
              <RotateCcw className="h-4 w-4" />
            </Button>
          </div>
        </form>

        {isStreaming && (
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <Loader2 className="h-4 w-4 animate-spin" />
            Streaming response...
          </div>
        )}
      </CardContent>
    </Card>
  )
}
