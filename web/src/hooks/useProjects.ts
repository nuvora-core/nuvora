import { listProjects } from '../api/projects'
import { useApiList } from './useApiList'

export function useProjects() {
  const {
    items: projects,
    loading,
    error,
    reload,
    refresh,
  } = useApiList(listProjects)

  return { projects, loading, error, reload, refresh }
}
