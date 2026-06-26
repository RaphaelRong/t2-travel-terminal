import { useEffect, useMemo, useRef, useState } from 'react'
import * as d3 from 'd3'
import * as topojson from 'topojson-client'
import type { GeometryCollection, Topology } from 'topojson-specification'
import type { CityMarket, CityMetricKey } from '../data/cityMarket'
import { getMetricValue } from '../data/cityMarket'

interface CityGlobeProps {
  cities: CityMarket[]
  selectedCityId: string
  metric: CityMetricKey
  onSelectCity: (cityId: string) => void
}

interface ProjectedPoint {
  city: CityMarket
  x: number
  y: number
  visible: boolean
  value: number
}

type Rotation = [number, number]
type DragState = {
  x: number
  y: number
  rotation: Rotation
}

const MIN_ZOOM = 0.9
const MAX_ZOOM = 1.85
const VIEWBOX = { width: 920, height: 760 }
const CENTER: [number, number] = [VIEWBOX.width / 2, VIEWBOX.height / 2 + 6]
const BASE_SCALE = 338

function metricColor(value: number) {
  if (value >= 88) return '#22c55e'
  if (value >= 78) return '#06b6d4'
  if (value >= 68) return '#d97706'
  return '#ef4444'
}

function clampLatitude(value: number) {
  return Math.max(-70, Math.min(70, value))
}

export function CityGlobe({ cities, selectedCityId, metric, onSelectCity }: CityGlobeProps) {
  const [land, setLand] = useState<GeoJSON.Feature | GeoJSON.FeatureCollection | null>(null)
  const [rotation, setRotation] = useState<Rotation>([-20, -18])
  const [zoom, setZoom] = useState(1.12)
  const [isPaused, setPaused] = useState(false)
  const dragRef = useRef<DragState | null>(null)

  const selectedCity = cities.find((city) => city.id === selectedCityId) || cities[0]
  const selectedLon = selectedCity?.lon
  const selectedLat = selectedCity?.lat

  useEffect(() => {
    let cancelled = false

    d3.json<Topology<{ land: GeometryCollection }>>('/data/countries-50m.json')
      .then((topology) => {
        if (!topology || cancelled) return
        setLand(topojson.feature(topology, topology.objects.land))
      })
      .catch((error) => {
        console.error('Failed to load globe land data:', error)
      })

    return () => {
      cancelled = true
    }
  }, [])

  useEffect(() => {
    if (isPaused) return undefined

    let frame = 0
    let last = performance.now()

    const tick = (now: number) => {
      const delta = now - last
      last = now
      setRotation((current) => [current[0] + delta * 0.0035, current[1]])
      frame = requestAnimationFrame(tick)
    }

    frame = requestAnimationFrame(tick)
    return () => cancelAnimationFrame(frame)
  }, [isPaused])

  useEffect(() => {
    if (selectedLon === undefined || selectedLat === undefined) return
    setRotation([-selectedLon, clampLatitude(-selectedLat)])
  }, [selectedCityId, selectedLat, selectedLon])

  const projection = useMemo(
    () =>
      d3
        .geoOrthographic()
        .scale(BASE_SCALE * zoom)
        .translate(CENTER)
        .rotate(rotation)
        .precision(0.6)
        .clipAngle(90),
    [rotation, zoom],
  )

  const path = useMemo(() => d3.geoPath(projection), [projection])
  const spherePath = path({ type: 'Sphere' }) || ''
  const graticulePath = path(d3.geoGraticule10()) || ''
  const landPath = land ? path(land) || '' : ''

  const points: ProjectedPoint[] = useMemo(() => {
    const centerLonLat: [number, number] = [-rotation[0], -rotation[1]]

    return cities.map((city) => {
      const coords = projection([city.lon, city.lat]) || [-1000, -1000]
      return {
        city,
        x: coords[0],
        y: coords[1],
        visible: d3.geoDistance([city.lon, city.lat], centerLonLat) < Math.PI / 2,
        value: getMetricValue(city, metric),
      }
    })
  }, [cities, metric, projection, rotation])

  const selectedPoint = points.find((point) => point.city.id === selectedCityId)
  const adjustZoom = (next: number) => setZoom(Math.min(MAX_ZOOM, Math.max(MIN_ZOOM, next)))

  const handlePointerDown = (event: React.PointerEvent<SVGSVGElement>) => {
    event.currentTarget.setPointerCapture(event.pointerId)
    dragRef.current = {
      x: event.clientX,
      y: event.clientY,
      rotation,
    }
    setPaused(true)
  }

  const handlePointerMove = (event: React.PointerEvent<SVGSVGElement>) => {
    const drag = dragRef.current
    if (!drag) return

    const dx = event.clientX - drag.x
    const dy = event.clientY - drag.y
    setRotation([drag.rotation[0] + dx * 0.32, clampLatitude(drag.rotation[1] - dy * 0.22)])
  }

  const handlePointerUp = (event: React.PointerEvent<SVGSVGElement>) => {
    dragRef.current = null
    event.currentTarget.releasePointerCapture(event.pointerId)
    setPaused(false)
  }

  return (
    <div
      className="relative h-full min-h-[460px] w-full"
      onMouseEnter={() => setPaused(true)}
      onMouseLeave={() => {
        if (!dragRef.current) setPaused(false)
      }}
      onWheel={(event) => {
        event.preventDefault()
        adjustZoom(zoom + (event.deltaY < 0 ? 0.08 : -0.08))
      }}
    >
      <svg
        viewBox={`0 0 ${VIEWBOX.width} ${VIEWBOX.height}`}
        className="h-full w-full touch-none"
        role="img"
        aria-label="T2 city data globe"
        onPointerDown={handlePointerDown}
        onPointerMove={handlePointerMove}
        onPointerUp={handlePointerUp}
        onPointerCancel={handlePointerUp}
      >
        <defs>
          <radialGradient id="t2Ocean" cx="38%" cy="34%" r="68%">
            <stop offset="0%" stopColor="rgb(var(--color-highlight))" />
            <stop offset="54%" stopColor="rgb(var(--color-surface))" />
            <stop offset="100%" stopColor="rgb(var(--color-base))" />
          </radialGradient>
          <filter id="cityGlow" x="-80%" y="-80%" width="260%" height="260%">
            <feGaussianBlur stdDeviation="5" result="blur" />
            <feMerge>
              <feMergeNode in="blur" />
              <feMergeNode in="SourceGraphic" />
            </feMerge>
          </filter>
        </defs>

        <path d={spherePath} fill="url(#t2Ocean)" stroke="rgb(var(--color-border-bright))" strokeOpacity="0.48" strokeWidth="1.2" />
        <path d={spherePath} fill="none" stroke="rgb(var(--color-cyan))" strokeOpacity="0.16" strokeWidth="18" />
        <path d={graticulePath} fill="none" stroke="rgb(var(--color-border-bright))" strokeOpacity="0.22" strokeWidth="0.8" />

        {landPath && (
          <>
            <path d={landPath} fill="rgb(var(--color-positive))" fillOpacity="0.08" stroke="rgb(var(--color-base))" strokeOpacity="0.72" strokeWidth="3.2" />
            <path d={landPath} fill="none" stroke="rgb(var(--color-text-primary))" strokeOpacity="0.34" strokeWidth="1.5" />
            <path d={landPath} fill="none" stroke="rgb(var(--color-cyan))" strokeOpacity="0.2" strokeWidth="0.8" />
          </>
        )}

        {selectedPoint?.visible && (
          <path
            d={`M ${CENTER[0]} ${CENTER[1]} L ${selectedPoint.x.toFixed(1)} ${selectedPoint.y.toFixed(1)}`}
            stroke="rgb(var(--color-accent))"
            strokeOpacity="0.28"
            strokeWidth="1.5"
            strokeDasharray="4 6"
          />
        )}

        {points
          .filter((point) => point.visible)
          .sort((a, b) => a.value - b.value)
          .map((point) => {
            const isSelected = point.city.id === selectedCityId
            const size = 6 + point.value / 9
            return (
              <g
                key={point.city.id}
                transform={`translate(${point.x} ${point.y})`}
                opacity={isSelected ? 1 : 0.84}
                className="cursor-pointer"
                role="button"
                tabIndex={0}
                onPointerDown={(event) => {
                  event.stopPropagation()
                }}
                onClick={(event) => {
                  event.stopPropagation()
                  onSelectCity(point.city.id)
                }}
                onKeyDown={(event) => {
                  if (event.key === 'Enter' || event.key === ' ') {
                    event.preventDefault()
                    onSelectCity(point.city.id)
                  }
                }}
              >
                <circle r={size + 11} fill={metricColor(point.value)} opacity={isSelected ? 0.16 : 0.08} filter="url(#cityGlow)" />
                <circle r={size} fill={metricColor(point.value)} stroke="rgb(var(--color-base))" strokeWidth="2" />
                {isSelected && (
                  <circle r={size + 8} fill="none" stroke="rgb(var(--color-text-primary))" strokeOpacity="0.88" strokeWidth="2" />
                )}
                {(isSelected || point.value >= 88) && (
                  <text
                    x={size + 13}
                    y="4"
                    className="pointer-events-none fill-obsidian-text-primary font-mono text-[13px] font-semibold"
                  >
                    {point.city.city}
                  </text>
                )}
              </g>
            )
          })}
      </svg>

      <div className="absolute bottom-4 right-4 flex items-center gap-1 border border-obsidian-border-dim bg-obsidian-base/90 p-1 backdrop-blur">
        <button
          type="button"
          aria-label="Zoom out"
          onClick={() => adjustZoom(zoom - 0.12)}
          className="h-8 w-8 rounded border border-obsidian-border-dim bg-obsidian-surface font-mono text-sm text-obsidian-text-secondary transition hover:border-obsidian-accent hover:text-obsidian-accent"
        >
          -
        </button>
        <span className="w-12 text-center font-mono text-[11px] text-obsidian-text-tertiary">{Math.round(zoom * 100)}%</span>
        <button
          type="button"
          aria-label="Zoom in"
          onClick={() => adjustZoom(zoom + 0.12)}
          className="h-8 w-8 rounded border border-obsidian-border-dim bg-obsidian-surface font-mono text-sm text-obsidian-text-secondary transition hover:border-obsidian-accent hover:text-obsidian-accent"
        >
          +
        </button>
      </div>
    </div>
  )
}
