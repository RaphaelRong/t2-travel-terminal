import { useMemo, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { CityGlobe } from '../components/CityGlobe'
import { ThemeToggle } from '../components/ThemeToggle'
import { useAuthStore } from '../store/authStore'
import { cities, getMetricValue, metricLabels, regions, type CityMarket, type CityMetricKey, type Region } from '../data/cityMarket'

const metricOptions: CityMetricKey[] = ['demand', 'adr', 'channel', 'event']

function formatNumber(value: number) {
  return new Intl.NumberFormat('en-US').format(value)
}

function average(items: CityMarket[], metric: CityMetricKey) {
  if (items.length === 0) return 0
  return Math.round(items.reduce((sum, city) => sum + getMetricValue(city, metric), 0) / items.length)
}

export function HomePage() {
  const navigate = useNavigate()
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated)
  const [region, setRegion] = useState<Region | 'All'>('All')
  const [metric, setMetric] = useState<CityMetricKey>('demand')
  const [selectedCityId, setSelectedCityId] = useState('dubai')

  const visibleCities = useMemo(
    () => (region === 'All' ? cities : cities.filter((city) => city.region === region)),
    [region],
  )

  const selectedCity =
    visibleCities.find((city) => city.id === selectedCityId) ||
    visibleCities[0] ||
    cities.find((city) => city.id === selectedCityId) ||
    cities[0]

  const topCities = [...visibleCities].sort((a, b) => getMetricValue(b, metric) - getMetricValue(a, metric)).slice(0, 5)
  const totalRooms = visibleCities.reduce((sum, city) => sum + city.roomSupply, 0)
  const avgMetric = average(visibleCities, metric)

  const openProtected = (path: string) => {
    if (isAuthenticated) {
      navigate(path)
      return
    }
    navigate(`/login?redirect=${encodeURIComponent(path)}`)
  }

  return (
    <div className="min-h-screen overflow-hidden bg-obsidian-base text-obsidian-text-primary">
      <header className="border-b border-obsidian-border-dim bg-obsidian-raised/95 backdrop-blur">
        <div className="mx-auto flex h-14 max-w-7xl items-center justify-between px-4">
          <Link to="/" className="flex items-center gap-3 font-mono text-sm font-semibold tracking-tight text-obsidian-text-primary">
            <span className="h-2.5 w-2.5 rounded-full bg-obsidian-positive shadow-[0_0_18px_rgb(var(--color-positive))]" />
            <span>T2 — Travel Terminal</span>
          </Link>
          <div className="flex items-center gap-2">
            <ThemeToggle />
            {isAuthenticated ? (
              <button
                type="button"
                onClick={() => navigate('/playground')}
                className="rounded border border-obsidian-border-dim bg-obsidian-surface px-3 py-1.5 font-mono text-xs text-obsidian-text-secondary transition hover:border-obsidian-accent hover:text-obsidian-accent"
              >
                Playground
              </button>
            ) : (
              <Link
                to="/login?redirect=%2Fplayground"
                className="rounded border border-obsidian-accent bg-obsidian-accent/10 px-3 py-1.5 font-mono text-xs text-obsidian-accent transition hover:bg-obsidian-accent hover:text-white"
              >
                Login
              </Link>
            )}
          </div>
        </div>
      </header>

      <main className="mx-auto grid min-h-[calc(100vh-3.5rem)] max-w-7xl grid-cols-1 gap-4 px-4 py-4 lg:grid-cols-[320px_minmax(0,1fr)_330px]">
        <aside className="order-2 space-y-4 lg:order-1">
          <section className="border border-obsidian-border-dim bg-obsidian-surface/92 p-4">
            <p className="font-mono text-[11px] uppercase text-obsidian-accent">Open city signal</p>
            <h1 className="mt-3 font-mono text-3xl font-bold leading-tight tracking-tight text-obsidian-text-primary">
              Global hotel intelligence, mapped by city.
            </h1>
            <p className="mt-3 font-mono text-sm leading-6 text-obsidian-text-secondary">
              T2 uses the globe as the shared data surface for travel demand, ADR momentum,
              channel health, and event pressure. City-level signals are free while deeper
              datasets stay behind authenticated access.
            </p>
          </section>

          <section className="border border-obsidian-border-dim bg-obsidian-surface/92 p-4">
            <div className="mb-3 flex items-center justify-between">
              <p className="font-mono text-xs uppercase text-obsidian-text-tertiary">Metric layer</p>
              <span className="font-mono text-xs text-obsidian-accent">{avgMetric}/100</span>
            </div>
            <div className="grid grid-cols-2 gap-2">
              {metricOptions.map((option) => (
                <button
                  key={option}
                  type="button"
                  onClick={() => setMetric(option)}
                  className={`rounded border px-3 py-2 text-left font-mono text-xs transition ${
                    metric === option
                      ? 'border-obsidian-accent bg-obsidian-accent/10 text-obsidian-accent'
                      : 'border-obsidian-border-dim bg-obsidian-base text-obsidian-text-secondary hover:border-obsidian-border-med hover:text-obsidian-text-primary'
                  }`}
                >
                  {metricLabels[option]}
                </button>
              ))}
            </div>
          </section>

          <section className="border border-obsidian-border-dim bg-obsidian-surface/92 p-4">
            <label className="mb-2 block font-mono text-xs uppercase text-obsidian-text-tertiary" htmlFor="region-filter">
              Region
            </label>
            <select
              id="region-filter"
              value={region}
              onChange={(event) => {
                const next = event.target.value as Region | 'All'
                setRegion(next)
                const first = next === 'All' ? cities[0] : cities.find((city) => city.region === next)
                if (first) setSelectedCityId(first.id)
              }}
              className="w-full rounded border border-obsidian-border-dim bg-obsidian-base px-3 py-2 font-mono text-sm text-obsidian-text-primary outline-none focus:border-obsidian-accent"
            >
              {regions.map((item) => (
                <option key={item} value={item}>
                  {item}
                </option>
              ))}
            </select>
          </section>
        </aside>

        <section className="relative order-1 min-h-[560px] overflow-hidden border border-obsidian-border-dim bg-obsidian-raised lg:order-2">
          <div className="absolute left-4 top-4 z-10 border border-obsidian-border-dim bg-obsidian-base/88 px-3 py-2 backdrop-blur">
            <p className="font-mono text-[10px] uppercase text-obsidian-text-tertiary">Public layer</p>
            <p className="font-mono text-sm text-obsidian-text-primary">{metricLabels[metric]}</p>
          </div>
          <CityGlobe
            cities={visibleCities}
            selectedCityId={selectedCity.id}
            metric={metric}
            onSelectCity={setSelectedCityId}
          />
          <div className="pointer-events-none absolute inset-x-0 bottom-0 h-28 bg-gradient-to-t from-obsidian-base/85 to-transparent" />
        </section>

        <aside className="order-3 space-y-4">
          <section className="border border-obsidian-border-dim bg-obsidian-surface/92 p-4">
            <div className="flex items-start justify-between gap-3">
              <div>
                <p className="font-mono text-[11px] uppercase text-obsidian-text-tertiary">{selectedCity.region}</p>
                <h2 className="mt-1 font-mono text-2xl font-bold text-obsidian-text-primary">{selectedCity.city}</h2>
                <p className="font-mono text-xs text-obsidian-text-secondary">{selectedCity.country}</p>
              </div>
              <span className="rounded border border-obsidian-cyan/40 bg-obsidian-cyan/10 px-2 py-1 font-mono text-[10px] uppercase text-obsidian-cyan">
                {selectedCity.marketTier}
              </span>
            </div>

            <div className="mt-5 grid grid-cols-2 gap-2">
              <div className="border border-obsidian-border-dim bg-obsidian-base p-3">
                <p className="font-mono text-xl font-bold text-obsidian-accent">{getMetricValue(selectedCity, metric)}</p>
                <p className="font-mono text-[10px] uppercase text-obsidian-text-tertiary">{metricLabels[metric]}</p>
              </div>
              <div className="border border-obsidian-border-dim bg-obsidian-base p-3">
                <p className="font-mono text-xl font-bold text-obsidian-positive">{selectedCity.trend > 0 ? '+' : ''}{selectedCity.trend}%</p>
                <p className="font-mono text-[10px] uppercase text-obsidian-text-tertiary">30d trend</p>
              </div>
              <div className="border border-obsidian-border-dim bg-obsidian-base p-3">
                <p className="font-mono text-xl font-bold text-obsidian-text-primary">{formatNumber(selectedCity.hotelSupply)}</p>
                <p className="font-mono text-[10px] uppercase text-obsidian-text-tertiary">Hotels</p>
              </div>
              <div className="border border-obsidian-border-dim bg-obsidian-base p-3">
                <p className="font-mono text-xl font-bold text-obsidian-text-primary">{formatNumber(selectedCity.roomSupply)}</p>
                <p className="font-mono text-[10px] uppercase text-obsidian-text-tertiary">Rooms</p>
              </div>
            </div>

            <div className="mt-5 space-y-3">
              {metricOptions.map((option) => (
                <div key={option}>
                  <div className="mb-1 flex justify-between font-mono text-[11px] text-obsidian-text-tertiary">
                    <span>{metricLabels[option]}</span>
                    <span>{getMetricValue(selectedCity, option)}</span>
                  </div>
                  <div className="h-1.5 bg-obsidian-highlight">
                    <div
                      className="h-full bg-obsidian-accent"
                      style={{ width: `${getMetricValue(selectedCity, option)}%` }}
                    />
                  </div>
                </div>
              ))}
            </div>
          </section>

          <section className="border border-obsidian-border-dim bg-obsidian-surface/92 p-4">
            <div className="mb-3 flex items-center justify-between">
              <p className="font-mono text-xs uppercase text-obsidian-text-tertiary">Top cities</p>
              <span className="font-mono text-xs text-obsidian-text-dim">{visibleCities.length} markets</span>
            </div>
            <div className="space-y-2">
              {topCities.map((city, index) => (
                <button
                  key={city.id}
                  type="button"
                  onClick={() => setSelectedCityId(city.id)}
                  className={`flex w-full items-center justify-between rounded border px-3 py-2 text-left font-mono text-xs transition ${
                    city.id === selectedCity.id
                      ? 'border-obsidian-accent bg-obsidian-accent/10 text-obsidian-accent'
                      : 'border-obsidian-border-dim bg-obsidian-base text-obsidian-text-secondary hover:border-obsidian-border-med hover:text-obsidian-text-primary'
                  }`}
                >
                  <span>{index + 1}. {city.city}</span>
                  <span>{getMetricValue(city, metric)}</span>
                </button>
              ))}
            </div>
          </section>

          <section className="border border-obsidian-border-dim bg-obsidian-surface/92 p-4">
            <p className="font-mono text-xs uppercase text-obsidian-accent">Authenticated data</p>
            <p className="mt-2 font-mono text-sm leading-6 text-obsidian-text-secondary">
              Sign in to open project workspaces, subscription plans, team access, and future premium
              dimensions such as channel mix, comp-set depth, forecast intervals, and asset-level signals.
            </p>
            <div className="mt-4 grid grid-cols-2 gap-2">
              <button
                type="button"
                onClick={() => openProtected('/projects')}
                className="rounded border border-obsidian-accent bg-obsidian-accent px-3 py-2 font-mono text-xs font-semibold text-white transition hover:bg-obsidian-accent/90"
              >
                View data
              </button>
              <button
                type="button"
                onClick={() => openProtected('/plans')}
                className="rounded border border-obsidian-border-dim bg-obsidian-base px-3 py-2 font-mono text-xs text-obsidian-text-secondary transition hover:border-obsidian-accent hover:text-obsidian-accent"
              >
                Plans
              </button>
            </div>
          </section>

          <section className="grid grid-cols-2 gap-2">
            <div className="border border-obsidian-border-dim bg-obsidian-surface/92 p-3">
              <p className="font-mono text-lg font-bold text-obsidian-text-primary">{formatNumber(totalRooms)}</p>
              <p className="font-mono text-[10px] uppercase text-obsidian-text-tertiary">Indexed rooms</p>
            </div>
            <div className="border border-obsidian-border-dim bg-obsidian-surface/92 p-3">
              <p className="font-mono text-lg font-bold text-obsidian-text-primary">${selectedCity.revenueSignal}m</p>
              <p className="font-mono text-[10px] uppercase text-obsidian-text-tertiary">Revenue signal</p>
            </div>
          </section>
        </aside>
      </main>
    </div>
  )
}
