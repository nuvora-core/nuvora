import Sidebar from './components/Sidebar'
import { useRoute } from './hooks/useRoute'
import Home from './pages/Home'
import ProjectPage from './pages/ProjectPage'
import Projects from './pages/Projects'
import Settings from './pages/Settings'
import './App.css'

function App() {
  const { route, navigate } = useRoute()

  if (route.name === 'project') {
    return (
      <ProjectPage
        projectId={route.id}
        onBack={() => navigate({ name: 'projects' })}
      />
    )
  }

  return (
    <div className="app">
      <Sidebar route={route} onNavigate={navigate} />

      <main className="main">
        {route.name === 'projects' && <Projects />}
        {route.name === 'settings' && <Settings />}
        {route.name === 'home' && <Home />}
      </main>
    </div>
  )
}

export default App
