import { useCallback, useEffect, useState } from 'react'

// useApiList loads a list from the API and keeps it in state, with
// reload (shows the loading state) and refresh (silent) helpers.
export function useApiList<T>(fetchItems: (signal?: AbortSignal) => Promise<T[]>) {
  const [items, setItems] = useState<T[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(
    async (signal?: AbortSignal) => {
      try {
        const loaded = await fetchItems(signal)

        setItems(loaded)
        setError(null)
        setLoading(false)
      } catch (err) {
        if (err instanceof DOMException && err.name === 'AbortError') {
          return
        }

        setError(err instanceof Error ? err.message : 'failed to load data')
        setLoading(false)
      }
    },
    [fetchItems],
  )

  useEffect(() => {
    const controller = new AbortController()

    // The initial fetch synchronizes with the API; the state updates
    // happen after the request resolves, not synchronously.
    // eslint-disable-next-line react/set-state-in-effect
    void load(controller.signal)

    return () => controller.abort()
  }, [load])

  const refresh = useCallback(() => {
    void load()
  }, [load])

  const reload = useCallback(() => {
    setLoading(true)
    void load()
  }, [load])

  return { items, setItems, loading, error, reload, refresh }
}
