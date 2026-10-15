import { useCallback, useEffect, useState } from 'react'

export type Route =
  | { name: 'home' }
  | { name: 'projects' }
  | { name: 'settings' }
  | { name: 'project'; id: number }

const PROJECTS_HASH = '#/projects'
const SETTINGS_HASH = '#/settings'

function parseHash(): Route {
  const hash = window.location.hash

  if (hash === PROJECTS_HASH) {
    return { name: 'projects' }
  }

  if (hash === SETTINGS_HASH) {
    return { name: 'settings' }
  }

  const projectMatch = /^#\/project\/(\d+)$/.exec(hash)

  if (projectMatch !== null) {
    return { name: 'project', id: Number(projectMatch[1]) }
  }

  return { name: 'home' }
}

function hashFor(route: Route): string {
  switch (route.name) {
    case 'projects':
      return PROJECTS_HASH
    case 'settings':
      return SETTINGS_HASH
    case 'project':
      return `#/project/${route.id}`
    default:
      return ''
  }
}

export function useRoute() {
  const [route, setRoute] = useState<Route>(parseHash)

  useEffect(() => {
    const handleHashChange = () => setRoute(parseHash())

    window.addEventListener('hashchange', handleHashChange)

    return () => window.removeEventListener('hashchange', handleHashChange)
  }, [])

  const navigate = useCallback((next: Route) => {
    window.location.hash = hashFor(next)
  }, [])

  return { route, navigate }
}
