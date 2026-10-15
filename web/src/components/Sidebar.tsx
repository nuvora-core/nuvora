import { FolderKanban, Settings, SunMoon } from 'lucide-react'
import { useTheme } from '../hooks/useTheme'
import type { Route } from '../hooks/useRoute'
import logo from '../assets/logo.svg'

type SidebarProps = {
  route: Route
  onNavigate: (route: Route) => void
}

function Sidebar({ route, onNavigate }: SidebarProps) {
  const { toggleTheme } = useTheme()

  const projectsActive = route.name === 'projects' || route.name === 'project'
  const settingsActive = route.name === 'settings'

  return (
    <aside className="sidebar" aria-label="Main navigation">
      <button
        type="button"
        className="sidebar-brand"
        aria-label="Nuvora home"
        onClick={() => onNavigate({ name: 'home' })}
      >
        <img className="sidebar-brand-logo" src={logo} alt="" width={24} height={24} />
        <span className="sidebar-brand-name">Nuvora</span>
      </button>

      <nav className="sidebar-nav" aria-label="Primary">
        <button
          type="button"
          className={`sidebar-button${projectsActive ? ' is-active' : ''}`}
          aria-label="Projects"
          aria-current={projectsActive ? 'page' : undefined}
          onClick={() => onNavigate({ name: 'projects' })}
        >
          <FolderKanban size={18} strokeWidth={1.75} aria-hidden="true" />
          <span className="sidebar-button-label">Projects</span>
        </button>
      </nav>

      <div className="sidebar-footer">
        <button
          type="button"
          className="sidebar-button"
          aria-label="Theme"
          onClick={toggleTheme}
        >
          <SunMoon size={18} strokeWidth={1.75} aria-hidden="true" />
          <span className="sidebar-button-label">Theme</span>
        </button>

        <button
          type="button"
          className={`sidebar-button${settingsActive ? ' is-active' : ''}`}
          aria-label="Settings"
          aria-current={settingsActive ? 'page' : undefined}
          onClick={() => onNavigate({ name: 'settings' })}
        >
          <Settings size={18} strokeWidth={1.75} aria-hidden="true" />
          <span className="sidebar-button-label">Settings</span>
        </button>
      </div>
    </aside>
  )
}

export default Sidebar
