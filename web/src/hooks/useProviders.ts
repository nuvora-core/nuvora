import { useCallback } from 'react'
import { deleteProvider, listProviders } from '../api/providers'
import { useApiList } from './useApiList'

export function useProviders() {
  const {
    items: providers,
    setItems,
    loading,
    error,
    reload,
    refresh,
  } = useApiList(listProviders)

  const remove = useCallback(
    async (id: number) => {
      await deleteProvider(id)
      setItems((current) => current.filter((provider) => provider.id !== id))
    },
    [setItems],
  )

  return { providers, loading, error, reload, refresh, remove }
}
