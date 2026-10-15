import { useCallback, useEffect, useState } from 'react'
import { listProviders } from '../api/providers'
import type { Provider } from '../api/providers'

export function useProviders() {
  const [providers, setProviders] = useState<Provider[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(async (signal?: AbortSignal) => {
    try {
      const loaded = await listProviders(signal)

      setProviders(loaded)
      setError(null)
      setLoading(false)
    } catch (err) {
      if (err instanceof DOMException && err.name === 'AbortError') {
        return
      }

      setError(err instanceof Error ? err.message : 'failed to load providers')
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    const controller = new AbortController()

    // The initial fetch synchronizes with the providers API; the state
    // updates happen after the request resolves, not synchronously.
    // eslint-disable-next-line react/set-state-in-effect
    void load(controller.signal)

    return () => controller.abort()
  }, [load])

  const reload = useCallback(() => {
    setLoading(true)
    void load()
  }, [load])

  return { providers, loading, error, reload }
}
