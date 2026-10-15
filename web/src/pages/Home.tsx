import logo from '../assets/logo.svg'

function Home() {
  return (
    <div className="home">
      <img className="main-logo" src={logo} alt="" width={96} height={96} />
      <h1 className="main-title">Welcome to Nuvora</h1>
    </div>
  )
}

export default Home
