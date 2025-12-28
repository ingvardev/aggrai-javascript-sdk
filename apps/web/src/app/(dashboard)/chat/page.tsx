import { StreamingChat } from '@/components/streaming-chat'

export default function ChatPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Streaming Chat</h1>
        <p className="text-muted-foreground">
          Real-time AI responses with Server-Sent Events
        </p>
      </div>

      <StreamingChat />
    </div>
  )
}
