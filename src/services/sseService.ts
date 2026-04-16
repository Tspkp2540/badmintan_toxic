const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api'

export type SSEEventType = 'room_created' | 'room_updated' | 'scores_submitted' | 'court_updated' | 'connected'

export type SSEListener = (data: any) => void

/**
 * SSE connection manager for real-time updates.
 * Connects to backend SSE endpoint and dispatches events to listeners.
 */
export class SSEConnection {
  private eventSource: EventSource | null = null
  private listeners = new Map<SSEEventType, Set<SSEListener>>()
  private url: string
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null

  constructor(url: string) {
    this.url = url
  }

  connect() {
    this.disconnect()
    this.eventSource = new EventSource(this.url)

    const eventTypes: SSEEventType[] = ['room_created', 'room_updated', 'scores_submitted', 'court_updated', 'connected']
    for (const type of eventTypes) {
      this.eventSource.addEventListener(type, (e: MessageEvent) => {
        let data: any = {}
        try { data = JSON.parse(e.data) } catch { /* empty */ }
        this.emit(type, data)
      })
    }

    this.eventSource.onerror = () => {
      this.eventSource?.close()
      this.eventSource = null
      // Reconnect after 3 seconds
      this.reconnectTimer = setTimeout(() => this.connect(), 3000)
    }
  }

  on(event: SSEEventType, listener: SSEListener) {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, new Set())
    }
    this.listeners.get(event)!.add(listener)
  }

  off(event: SSEEventType, listener: SSEListener) {
    this.listeners.get(event)?.delete(listener)
  }

  private emit(event: SSEEventType, data: any) {
    this.listeners.get(event)?.forEach(fn => fn(data))
  }

  disconnect() {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    if (this.eventSource) {
      this.eventSource.close()
      this.eventSource = null
    }
    this.listeners.clear()
  }
}

/** Create an SSE connection for a specific court */
export function createCourtSSE(courtId: string): SSEConnection {
  const conn = new SSEConnection(`${API_BASE_URL}/events?courtId=${encodeURIComponent(courtId)}`)
  return conn
}

/** Create a global SSE connection for court list updates */
export function createGlobalSSE(): SSEConnection {
  const conn = new SSEConnection(`${API_BASE_URL}/events?scope=global`)
  return conn
}
