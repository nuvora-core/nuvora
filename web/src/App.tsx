import Sidebar from './components/Sidebar'
import { useRoute } from './hooks/useRoute'
import Home from './pages/Home'
import Settings from './pages/Settings'
import './App.css'

function App() {
  const { route, navigate } = useRoute()

  return (
    <div className="app">
      <Sidebar route={route} onNavigate={navigate} />

      <main className="main">
        {route === 'settings' ? <Settings /> : <Home />}
      </main>
    </div>
  )
}

export default App
