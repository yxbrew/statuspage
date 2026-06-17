import { useEffect, useMemo, useState } from 'react'
import './App.css'

type HealthPayload = {
  status?: string
  service?: string
  version?: string
}

const POLL_INTERVAL_MS = 12000

function App() {
  const [health, setHealth] = useState<HealthPayload | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [lastChecked, setLastChecked] = useState<string | null>(null)

  const fetchHealth = async () => {
    try {
      const response = await fetch('/api/v1/health', {
        headers: {
          Accept: 'application/json',
        },
      })

      const data = (await response.json()) as HealthPayload
      setHealth(data)
      setError(response.ok ? null : 'Backend returned a non-OK status')
      setLastChecked(new Date().toLocaleTimeString())
    } catch {
      setError('Unable to reach backend health endpoint')
      setHealth(null)
      setLastChecked(new Date().toLocaleTimeString())
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void fetchHealth()
    const timer = window.setInterval(() => {
      void fetchHealth()
    }, POLL_INTERVAL_MS)

    return () => {
      window.clearInterval(timer)
    }
  }, [])

  const systemState = useMemo(() => {
    if (loading) {
      return 'Checking'
    }
    if (error) {
      return 'Degraded'
    }
    if (health?.status?.toLowerCase() === 'ok') {
      return 'Operational'
    }
    return 'Unknown'
  }, [error, health?.status, loading])

  const statusTone = useMemo(() => {
    if (systemState === 'Operational') {
      return 'ok'
    }
    if (systemState === 'Checking') {
      return 'checking'
    }
    return 'bad'
  }, [systemState])

  return (
    <main className="dashboard">
      <header className="topbar">
        <div className="title-wrap">
          <p className="eyebrow">Status Monitor</p>
          <h1>Service Control Panel</h1>
        </div>
        <button className="refresh" type="button" onClick={() => void fetchHealth()}>
          Refresh
        </button>
      </header>

      <section className="hero-grid">
        <article className="card pulse">
          <p className="label">System State</p>
          <div className="state-row">
            <span className={`dot ${statusTone}`} aria-hidden="true"></span>
            <strong>{systemState}</strong>
          </div>
          <p className="muted">Polled every {Math.floor(POLL_INTERVAL_MS / 1000)}s</p>
        </article>

        <article className="card">
          <p className="label">Endpoint</p>
          <p className="mono">/api/v1/health</p>
          <p className="muted">Proxy target: http://localhost:8080</p>
        </article>

        <article className="card">
          <p className="label">Last Check</p>
          <p className="mono">{lastChecked ?? 'Not yet checked'}</p>
          <p className="muted">Local browser time</p>
        </article>
      </section>

      <section className="table-card">
        <div className="table-head">
          <h2>Backend Health</h2>
          <span className={`badge ${statusTone}`}>{systemState}</span>
        </div>
        <div className="rows">
          <div className="row">
            <span>service</span>
            <code>{health?.service ?? '-'}</code>
          </div>
          <div className="row">
            <span>status</span>
            <code>{health?.status ?? (loading ? 'loading' : '-')}</code>
          </div>
          <div className="row">
            <span>version</span>
            <code>{health?.version ?? '-'}</code>
          </div>
          {error ? (
            <div className="row row-error">
              <span>error</span>
              <code>{error}</code>
            </div>
          ) : null}
        </div>
      </section>

      <footer className="footnote">
        Inspired by AIwatch-style monitoring dashboards, adapted for your backend.
      </footer>
    </main>
  )
}

export default App
