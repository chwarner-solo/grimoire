import { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import {
  useFactionsQuery, useCreateFactionMutation, FactionsDocument, FactionSummary,
} from '../../../graphql/generated'
import { StatusBadge } from '../../../components/ui/StatusBadge'
import { Spinner } from '../../../components/ui/Spinner'
import { EmptyState } from '../../../components/ui/EmptyState'
import { ErrorMessage } from '../../../components/ui/ErrorMessage'
import { Modal } from '../../../components/ui/Modal'

export default function FactionsPage() {
  const { gameId } = useParams<{ gameId: string }>()
  const { data, loading, error } = useFactionsQuery({ variables: { gameId: gameId! } })
  const [showCreate, setShowCreate] = useState(false)

  if (loading) return <Spinner />
  if (error) return <div className="p-6"><ErrorMessage message={error.message} /></div>

  const factions = data?.factions ?? []

  return (
    <div className="p-6 max-w-3xl">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-xl font-bold text-slate-100">Factions</h1>
        <button onClick={() => setShowCreate(true)} className="btn-primary">+ Add Faction</button>
      </div>

      {factions.length === 0 ? (
        <EmptyState
          title="No factions yet"
          description="Factions add political depth — guilds, cults, governments, and cabals."
          action={<button onClick={() => setShowCreate(true)} className="btn-primary">Create Faction</button>}
        />
      ) : (
        <div className="space-y-3">
          {factions.map(f => (
            <FactionCard key={f.id} faction={f} gameId={gameId!} />
          ))}
        </div>
      )}

      {showCreate && (
        <CreateFactionModal gameId={gameId!} onClose={() => setShowCreate(false)} />
      )}
    </div>
  )
}

function FactionCard({ faction, gameId }: { faction: FactionSummary; gameId: string }) {
  const navigate = useNavigate()
  return (
    <div className="bg-slate-800 border border-slate-700 rounded-lg p-4">
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center gap-3">
          <h2 className="font-semibold text-slate-200">{faction.name}</h2>
          <StatusBadge status={faction.status} />
        </div>
        <button
          onClick={() => navigate(`/world-building/games/${gameId}/factions/${faction.id}`)}
          className="text-violet-400 hover:text-violet-300 text-sm"
        >
          →
        </button>
      </div>
      <div className="text-sm text-slate-400 flex gap-4">
        <span>{faction.memberCount} member{faction.memberCount !== 1 ? 's' : ''}</span>
        <span>{faction.standingLevelCount} standing levels</span>
      </div>
      {faction.allies.length > 0 && (
        <p className="text-xs text-slate-500 mt-1">Allied with: {faction.allies.map(a => a.name).join(', ')}</p>
      )}
      {faction.enemies.length > 0 && (
        <p className="text-xs text-slate-500">At war with: {faction.enemies.map(e => e.name).join(', ')}</p>
      )}
    </div>
  )
}

function CreateFactionModal({ gameId, onClose }: { gameId: string; onClose: () => void }) {
  const navigate = useNavigate()
  const [name, setName] = useState('')
  const [createFaction, { loading, error }] = useCreateFactionMutation({
    refetchQueries: [{ query: FactionsDocument, variables: { gameId } }],
  })

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!name.trim()) return
    const result = await createFaction({ variables: { input: { gameId, name: name.trim() } } })
    const id = result.data?.createFaction?.id
    if (id) {
      onClose()
      navigate(`/world-building/games/${gameId}/factions/${id}`)
    }
  }

  return (
    <Modal title="Create Faction" onClose={onClose}>
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="field-label">Faction Name</label>
          <input autoFocus value={name} onChange={e => setName(e.target.value)} className="input w-full" />
        </div>
        {error && <ErrorMessage message={error.message} />}
        <div className="flex justify-end gap-3">
          <button type="button" onClick={onClose} className="btn-ghost">Cancel</button>
          <button type="submit" disabled={loading || !name.trim()} className="btn-primary">Create</button>
        </div>
      </form>
    </Modal>
  )
}
