import { useState } from 'react'
import { useParams } from 'react-router-dom'
import {
  useLocationsQuery, useLocationQuery, useCreateLocationMutation, useActivateLocationMutation,
  useAddSceneMutation,
  LocationsDocument, LocationDocument,
  LocationSummary, LocationDetail as LocationDetailType, EntityStatus, LocationType,
} from '../../../graphql/generated'
import { StatusBadge } from '../../../components/ui/StatusBadge'
import { Spinner } from '../../../components/ui/Spinner'
import { EmptyState } from '../../../components/ui/EmptyState'
import { ErrorMessage } from '../../../components/ui/ErrorMessage'

const LOCATION_ICONS: Record<LocationType, string> = {
  [LocationType.World]:      '🌍',
  [LocationType.Region]:     '🗺️',
  [LocationType.Settlement]: '🏘️',
  [LocationType.Building]:   '🏛️',
  [LocationType.Scene]:      '🚪',
}

export default function LocationsPage() {
  const { gameId } = useParams<{ gameId: string }>()
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const { data, loading, error } = useLocationsQuery({ variables: { gameId: gameId! } })

  if (loading) return <Spinner />
  if (error) return <div className="p-6"><ErrorMessage message={error.message} /></div>

  const locations = data?.locations ?? []
  const roots = buildTree(locations)

  return (
    <div className="flex h-full">
      <div className="w-80 border-r border-slate-700 flex flex-col">
        <div className="flex items-center justify-between px-4 py-3 border-b border-slate-700">
          <h1 className="font-semibold text-slate-100">Locations</h1>
          <AddLocationButton gameId={gameId!} parentId={null} label="+ Add" />
        </div>

        <div className="flex-1 overflow-y-auto px-2 py-2">
          {roots.length === 0 ? (
            <EmptyState title="No locations yet" description="Start with the world or a region." />
          ) : (
            roots.map(node => (
              <LocationTreeNode
                key={node.id}
                node={node}
                gameId={gameId!}
                selectedId={selectedId}
                onSelect={setSelectedId}
                depth={0}
              />
            ))
          )}
        </div>
      </div>

      <div className="flex-1 overflow-y-auto">
        {selectedId ? (
          <LocationDetailPanel locationId={selectedId} gameId={gameId!} />
        ) : (
          <div className="flex items-center justify-center h-full text-slate-500 text-sm">
            Select a location to view or edit it.
          </div>
        )}
      </div>
    </div>
  )
}

type TreeNode = LocationSummary & { children: TreeNode[] }

function buildTree(locations: LocationSummary[]): TreeNode[] {
  const map = new Map<string, TreeNode>()
  locations.forEach(l => map.set(l.id, { ...l, children: [] }))
  const roots: TreeNode[] = []
  locations.forEach(l => {
    const node = map.get(l.id)!
    if (l.parentId && map.has(l.parentId)) {
      map.get(l.parentId)!.children.push(node)
    } else {
      roots.push(node)
    }
  })
  return roots
}

function LocationTreeNode({
  node, gameId, selectedId, onSelect, depth,
}: {
  node: TreeNode; gameId: string; selectedId: string | null
  onSelect: (id: string) => void; depth: number
}) {
  const [expanded, setExpanded] = useState(true)
  const isSelected = node.id === selectedId
  const isArchived = node.status === EntityStatus.Archived

  return (
    <div>
      <div
        className={`flex items-center gap-1 px-2 py-1.5 rounded cursor-pointer group
          ${isSelected ? 'bg-violet-900/40 text-violet-300' : 'hover:bg-slate-700/50 text-slate-300'}
          ${isArchived ? 'opacity-40' : ''}
        `}
        style={{ paddingLeft: `${8 + depth * 16}px` }}
        onClick={() => onSelect(node.id)}
      >
        {node.children.length > 0 && (
          <button
            onClick={e => { e.stopPropagation(); setExpanded(!expanded) }}
            className="w-4 text-center text-slate-500 hover:text-slate-300 text-xs"
          >
            {expanded ? '▼' : '▶'}
          </button>
        )}
        {node.children.length === 0 && <span className="w-4" />}

        <span className="text-base">{LOCATION_ICONS[node.locationType]}</span>
        <span className="flex-1 text-sm truncate">{node.name}</span>

        {node.status !== EntityStatus.Active && node.status !== EntityStatus.New && (
          <span className="text-xs text-slate-500">[{node.status}]</span>
        )}

        <div className="opacity-0 group-hover:opacity-100" onClick={e => e.stopPropagation()}>
          <AddLocationButton gameId={gameId} parentId={node.id} label="+" />
        </div>
      </div>

      {expanded && node.children.map(child => (
        <LocationTreeNode
          key={child.id}
          node={child}
          gameId={gameId}
          selectedId={selectedId}
          onSelect={onSelect}
          depth={depth + 1}
        />
      ))}
    </div>
  )
}

function AddLocationButton({ gameId, parentId, label }: { gameId: string; parentId: string | null; label: string }) {
  const [open, setOpen] = useState(false)
  const [name, setName] = useState('')
  const [locationType, setLocationType] = useState<LocationType>(LocationType.Settlement)
  const [createLocation, { loading }] = useCreateLocationMutation({
    refetchQueries: [{ query: LocationsDocument, variables: { gameId } }],
  })

  if (!open) {
    return (
      <button
        onClick={() => setOpen(true)}
        className="text-xs text-slate-500 hover:text-violet-400 px-1 py-0.5 rounded"
      >
        {label}
      </button>
    )
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!name.trim()) return
    await createLocation({ variables: { input: { gameId, name: name.trim(), locationType, ...(parentId ? { parentId } : {}) } } })
    setOpen(false)
    setName('')
  }

  return (
    <form onSubmit={handleSubmit} onClick={e => e.stopPropagation()} className="flex items-center gap-1 p-1">
      <input
        autoFocus
        value={name}
        onChange={e => setName(e.target.value)}
        placeholder="Location name"
        className="input text-xs w-32"
      />
      <select
        value={locationType}
        onChange={e => setLocationType(e.target.value as LocationType)}
        className="input text-xs"
      >
        {Object.values(LocationType).map(t => <option key={t} value={t}>{t}</option>)}
      </select>
      <button type="submit" disabled={loading} className="text-xs text-violet-400 hover:text-violet-300">✓</button>
      <button type="button" onClick={() => setOpen(false)} className="text-xs text-slate-500">✕</button>
    </form>
  )
}

function LocationDetailPanel({ locationId, gameId }: { locationId: string; gameId: string }) {
  const { data, loading, error } = useLocationQuery({ variables: { id: locationId, gameId } })
  const [activateLocation] = useActivateLocationMutation({
    refetchQueries: [{ query: LocationDocument, variables: { id: locationId, gameId } }],
  })

  if (loading) return <Spinner />
  if (error) return <div className="p-6"><ErrorMessage message={error.message} /></div>
  if (!data?.location) return null

  const loc = data.location
  const isActive = loc.status === EntityStatus.Active

  return (
    <div className="p-6 max-w-xl space-y-6">
      <div className="flex items-center gap-3">
        <h2 className="text-xl font-bold text-slate-100">{loc.name}</h2>
        <StatusBadge status={loc.status} />
        <span className="text-xs text-slate-500 uppercase">{loc.locationType}</span>
      </div>

      {loc.parentLocation && (
        <p className="text-sm text-slate-500">
          Within: {LOCATION_ICONS[loc.parentLocation.locationType]} {loc.parentLocation.name}
        </p>
      )}

      <ScenesSection location={loc} gameId={gameId} />
      <ConnectionsSection location={loc} />

      {!isActive && (
        <button
          onClick={() => activateLocation({ variables: { input: { locationId, gameId } } })}
          className="btn-primary w-full"
        >
          Activate Location ▶
        </button>
      )}
    </div>
  )
}

function ScenesSection({ location, gameId }: { location: LocationDetailType; gameId: string }) {
  const [adding, setAdding] = useState(false)
  const [sceneName, setSceneName] = useState('')
  const [addScene, { loading }] = useAddSceneMutation({
    refetchQueries: [{ query: LocationDocument, variables: { id: location.id, gameId } }],
  })

  async function handleAdd(e: React.FormEvent) {
    e.preventDefault()
    if (!sceneName.trim()) return
    await addScene({ variables: { input: { locationId: location.id, gameId, name: sceneName.trim() } } })
    setSceneName('')
    setAdding(false)
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-2">
        <h3 className="text-sm font-medium text-slate-300">Scenes</h3>
        <button onClick={() => setAdding(true)} className="text-xs text-violet-400 hover:text-violet-300">
          + Scene
        </button>
      </div>
      <div className="bg-slate-800 border border-slate-700 rounded-lg divide-y divide-slate-700">
        {location.scenes.map(scene => (
          <div key={scene.id} className="px-3 py-2 text-sm text-slate-300">{scene.name}</div>
        ))}
        {adding && (
          <form onSubmit={handleAdd} className="flex items-center gap-2 px-3 py-2">
            <input
              autoFocus
              value={sceneName}
              onChange={e => setSceneName(e.target.value)}
              placeholder="Scene name"
              className="input text-sm flex-1"
            />
            <button type="submit" disabled={loading} className="text-xs text-violet-400">✓</button>
            <button type="button" onClick={() => setAdding(false)} className="text-xs text-slate-500">✕</button>
          </form>
        )}
        {location.scenes.length === 0 && !adding && (
          <div className="px-3 py-3 text-sm text-slate-500">No scenes yet.</div>
        )}
      </div>
    </div>
  )
}

function ConnectionsSection({ location }: { location: LocationDetailType }) {
  const directionLabel = { OUTBOUND: '→', INBOUND: '←', BOTH: '↔' }

  return (
    <div>
      <h3 className="text-sm font-medium text-slate-300 mb-2">Travel Connections</h3>
      <div className="bg-slate-800 border border-slate-700 rounded-lg divide-y divide-slate-700">
        {location.connections.map((conn, i) => (
          <div key={i} className="px-3 py-2 text-sm text-slate-300 flex items-center gap-2">
            <span className="text-slate-500">{directionLabel[conn.direction]}</span>
            <span>{conn.toLocation.name}</span>
          </div>
        ))}
        {location.connections.length === 0 && (
          <div className="px-3 py-3 text-sm text-slate-500">No connections yet.</div>
        )}
      </div>
    </div>
  )
}
