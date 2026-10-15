import { Settings, SunMoon } from 'lucide-react'
import { useTheme } from '../hooks/useTheme'
import type { Route } from '../hooks/useRoute'
import logo from '../assets/logo.svg'

type SidebarProps = {
  route: Route
  onNavigate: (route: Route) => void
}

function Sidebar({ route, onNavigate }: SidebarProps) {
  const { toggleTheme } = useTheme()

  return (
    <aside className="sidebar" aria-label="Main navigation">
      <button
        type="button"
        className="sidebar-brand"
        aria-label="Nuvora home"
        onClick={() => onNavigate('home')}
      >
        <img className="sidebar-brand-logo" src={logo} alt="" width={24} height={24} />
        <span className="sidebar-brand-name">Nuvora</span>
      </button>

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
          className={`sidebar-button${route === 'settings' ? ' is-active' : ''}`}
          aria-label="Settings"
          aria-current={route === 'settings' ? 'page' : undefined}
          onClick={() => onNavigate('settings')}
        >
          <Settings size={18} strokeWidth={1.75} aria-hidden="true" />
          <span className="sidebar-button-label">Settings</span>
        </button>
      </div>
    </aside>
  )
}

export default Sidebar
