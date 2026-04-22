import { useEffect, useRef, useCallback } from 'react'

type MessageHandler = (channel: string, data: unknown) => void

export function useWebSocket(onMessage: MessageHandler) {
  const wsRef = useRef<WebSocket | null>(null)
  const handlersRef = useRef(onMessage)
  handlersRef.current = onMessage

  useEffect(() => {
    const apiBase = import.meta.env.VITE_API_URL
    let wsURL: string
    if (apiBase) {
      wsURL = apiBase.replace(/^http/, 'ws') + '/ws'
    } else {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      wsURL = `${protocol}//${window.location.host}/ws`
    }
    const ws = new WebSocket(wsURL)
    wsRef.current = ws

    ws.onmessage = (e) => {
      try {
        const envelope = JSON.parse(e.data) as { channel: string; data: unknown }
        handlersRef.current(envelope.channel, envelope.data)
      } catch {
        // ignore malformed
      }
    }
    ws.onerror = (e) => console.error('ws error', e)

    return () => {
      ws.close()
      wsRef.current = null
    }
  }, [])

  const subscribe = useCallback((channels: string[]) => {
    const ws = wsRef.current
    if (!ws) return
    const send = () => ws.send(JSON.stringify({ subscribe: channels }))
    if (ws.readyState === WebSocket.OPEN) {
      send()
    } else {
      ws.addEventListener('open', send, { once: true })
    }
  }, [])

  return { subscribe }
}
