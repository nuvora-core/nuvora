import { useState } from 'react'
import type { FormEvent } from 'react'
import { Plug } from 'lucide-react'
import { connectProvider } from '../api/providers'
import ProviderLogo from '../components/ProviderLogo'
import { useProviders } from '../hooks/useProviders'
import { PROVIDER_CATALOG, getProviderMeta } from '../lib/providers'
import type { ProviderMeta } from '../lib/providers'
import './Settings.css'

type ConnectCardProps = {
  provider: ProviderMeta
  onConnected: () => void
}

function ConnectCard({ provider, onConnected }: ConnectCardProps) {
  const [apiKey, setApiKey] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    const key = apiKey.trim()

    if (key === '') {
      setError('API key is required')

      return
    }

    setSubmitting(true)
    setError(null)

    try {
      await connectProvider(provider.id, key)
      setApiKey('')
      onConnected()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'failed to connect provider')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <form className="provider-card" onSubmit={handleSubmit}>
      <div className="provider-card-header">
        <span className="provider-tile">
          <ProviderLogo provider={provider} />
        </span>
        <div className="provider-card-info">
          <h3 className="provider-name">{provider.name}</h3>
          <p className="provider-description">{provider.description}</p>
        </div>
      </div>

      <div className="provider-card-body">
        <label className="field">
          <span className="field-label">API key</span>
          <input
            className="input"
            type="password"
            value={apiKey}
            onChange={(event) => setApiKey(event.target.value)}
            placeholder={`${provider.name} API key`}
            autoComplete="off"
            spellCheck={false}
          />
        </label>

        {error !== null && (
          <p className="provider-error" role="alert">
            {error}
          </p>
        )}

        <button type="submit" className="button-primary" disabled={submitting}>
          {submitting ? 'Connecting...' : 'Connect'}
        </button>
      </div>
    </form>
  )
}

function Settings() {
  const { providers, loading, error, reload } = useProviders()

  return (
    <div className="settings">
      <header className="settings-header">
        <h1 className="settings-title">Settings</h1>
        <p className="settings-subtitle">
          Manage the providers powering your models.
        </p>
      </header>

      <section className="settings-section">
        <h2 className="settings-section-title">Your Providers</h2>

        {loading && <p className="muted">Loading providers...</p>}

        {!loading && error !== null && (
          <div className="settings-error" role="alert">
            <p className="settings-error-message">{error}</p>
            <button type="button" className="button-secondary" onClick={reload}>
              Try again
            </button>
          </div>
        )}

        {!loading && error === null && providers.length === 0 && (
          <p className="muted">
            No providers connected yet. Connect one below to get started.
          </p>
        )}

        {!loading && error === null && providers.length > 0 && (
          <div className="provider-grid">
            {providers.map((provider) => {
              const meta = getProviderMeta(provider.name)

              return (
                <article className="provider-card" key={provider.base_url}>
                  <div className="provider-card-header">
                    <span className="provider-tile">
                      {meta !== undefined ? (
                        <ProviderLogo provider={meta} />
                      ) : (
                        <Plug size={22} strokeWidth={1.75} aria-hidden="true" />
                      )}
                    </span>
                    <div className="provider-card-info">
                      <h3 className="provider-name">
                        {meta?.name ?? provider.name}
                      </h3>
                      <p className="provider-base-url">{provider.base_url}</p>
                    </div>
                  </div>

                  <span className="provider-status">
                    <span className="provider-status-dot" aria-hidden="true" />
                    Connected
                  </span>
                </article>
              )
            })}
          </div>
        )}
      </section>

      <section className="settings-section">
        <h2 className="settings-section-title">Connect provider</h2>

        <div className="provider-grid">
          {PROVIDER_CATALOG.map((provider) => (
            <ConnectCard
              key={provider.id}
              provider={provider}
              onConnected={reload}
            />
          ))}
        </div>
      </section>
    </div>
  )
}

export default Settings
