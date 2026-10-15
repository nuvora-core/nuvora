import { useCallback, useEffect, useState } from 'react'

export type Route = 'home' | 'settings'

const ROUTE_HASHES: Record<Route, string> = {
  home: '',
  settings: '#/settings',
}

function getRoute(): Route {
  return window.location.hash === ROUTE_HASHES.settings ? 'settings' : 'home'
}

export function useRoute() {
  const [route, setRoute] = useState<Route>(getRoute)

  useEffect(() => {
    const handleHashChange = () => setRoute(getRoute())

    window.addEventListener('hashchange', handleHashChange)

    return () => window.removeEventListener('hashchange', handleHashChange)
  }, [])

  const navigate = useCallback((next: Route) => {
    window.location.hash = ROUTE_HASHES[next]
  }, [])

  return { route, navigate }
}
