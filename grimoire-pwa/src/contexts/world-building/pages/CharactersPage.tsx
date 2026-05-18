import { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import {
  useNpcsQuery, usePlayerCharactersQuery, useCreateNpcMutation, useCreatePlayerCharacterMutation,
  NpcsDocument, PlayerCharactersDocument, NpcSummary, PlayerCharacterSummary,
} from '../../../graphql/generated'
import { StatusBadge } from '../../../components/ui/StatusBadge'
import { Spinner } from '../../../components/ui/Spinner'
import { EmptyState } from '../../../components/ui/EmptyState'
import { ErrorMessage } from '../../../components/ui/ErrorMessage'
import { Modal } from '../../../components/ui/Modal'

type Tab = 'npcs' | 'pcs'

export default function CharactersPage() {
  const { gameId } = useParams<{ gameId: string }>()
  const [tab, setTab] = useState<Tab>('npcs')

  return (
    <div className="p-6 max-w-3xl">
      <h1 className="text-xl font-bold text-slate-100 mb-4">Characters</h1>

      <div className="flex border-b border-slate-700 mb-4">
        {(['npcs', 'pcs'] as Tab[]).map(t => (
          <button
            key={t}
            onClick={() => setTab(t)}
            className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors -mb-px ${
              tab === t ? 'border-violet-500 text-violet-300' : 'border-transparent text-slate-400 hover:text-slate-200'
            }`}
          >
            {t === 'npcs' ? 'NPCs' : 'Player Characters'}
          </button>
        ))}
      </div>

      {tab === 'npcs' ? <NpcTab gameId={gameId!} /> : <PcTab gameId={gameId!} />}
    </div>
  )
}

function NpcTab({ gameId }: { gameId: string }) {
  const [showCreate, setShowCreate] = useState(false)
  const { data, loading, error } = useNpcsQuery({ variables: { gameId } })

  if (loading) return <Spinner />
  if (error) return <ErrorMessage message={error.message} />

  const npcs = data?.npcs ?? []

  return (
    <>
      <div className="flex justify-end mb-3">
        <button onClick={() => setShowCreate(true)} className="btn-primary text-sm">+ NPC</button>
      </div>

      {npcs.length === 0 ? (
        <EmptyState title="No NPCs yet" description="Create NPCs to populate your world." />
      ) : (
        <div className="bg-slate-800 border border-slate-700 rounded-lg divide-y divide-slate-700">
          {npcs.map(npc => <NpcRow key={npc.id} npc={npc} gameId={gameId} />)}
        </div>
      )}

      {showCreate && <CreateNpcModal gameId={gameId} onClose={() => setShowCreate(false)} />}
    </>
  )
}

function NpcRow({ npc, gameId }: { npc: NpcSummary; gameId: string }) {
  const navigate = useNavigate()
  return (
    <button
      onClick={() => navigate(`/world-building/games/${gameId}/characters/npcs/${npc.id}`)}
      className="w-full flex items-center justify-between px-4 py-3 hover:bg-slate-700/50 transition-colors text-left"
    >
      <span className="font-medium text-slate-200">{npc.name}</span>
      <div className="flex items-center gap-3">
        {!npc.playerVisible && <span className="text-xs text-slate-500">GM-only</span>}
        <StatusBadge status={npc.status} />
      </div>
    </button>
  )
}

function PcTab({ gameId }: { gameId: string }) {
  const [showCreate, setShowCreate] = useState(false)
  const { data, loading, error } = usePlayerCharactersQuery({ variables: { gameId } })

  if (loading) return <Spinner />
  if (error) return <ErrorMessage message={error.message} />

  const pcs = data?.playerCharacters ?? []

  return (
    <>
      <div className="flex justify-end mb-3">
        <button onClick={() => setShowCreate(true)} className="btn-primary text-sm">+ Player Character</button>
      </div>

      {pcs.length === 0 ? (
        <EmptyState title="No player characters yet" description="Add player characters to campaigns in the Campaign Setup wizard." />
      ) : (
        <div className="bg-slate-800 border border-slate-700 rounded-lg divide-y divide-slate-700">
          {pcs.map(pc => <PcRow key={pc.id} pc={pc} />)}
        </div>
      )}

      {showCreate && <CreatePcModal gameId={gameId} onClose={() => setShowCreate(false)} />}
    </>
  )
}

function PcRow({ pc }: { pc: PlayerCharacterSummary }) {
  return (
    <div className="flex items-center justify-between px-4 py-3">
      <div>
        <span className="font-medium text-slate-200">{pc.name}</span>
        {pc.ownerPlayerId && (
          <span className="text-xs text-slate-500 ml-2">Player: {pc.ownerPlayerId}</span>
        )}
      </div>
      <StatusBadge status={pc.status} />
    </div>
  )
}

function CreateNpcModal({ gameId, onClose }: { gameId: string; onClose: () => void }) {
  const navigate = useNavigate()
  const [name, setName] = useState('')
  const [createNPC, { loading, error }] = useCreateNpcMutation({
    refetchQueries: [{ query: NpcsDocument, variables: { gameId } }],
  })

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!name.trim()) return
    const result = await createNPC({ variables: { input: { gameId, name: name.trim() } } })
    const id = result.data?.createNPC?.id
    if (id) {
      onClose()
      navigate(`/world-building/games/${gameId}/characters/npcs/${id}`)
    }
  }

  return (
    <Modal title="Create NPC" onClose={onClose}>
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="field-label">Name</label>
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

function CreatePcModal({ gameId, onClose }: { gameId: string; onClose: () => void }) {
  const [name, setName] = useState('')
  const [ownerPlayerId, setOwnerPlayerId] = useState('')
  const [createPC, { loading, error }] = useCreatePlayerCharacterMutation({
    refetchQueries: [{ query: PlayerCharactersDocument, variables: { gameId } }],
  })

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!name.trim()) return
    await createPC({ variables: { input: { gameId, name: name.trim(), ...(ownerPlayerId.trim() ? { ownerPlayerId: ownerPlayerId.trim() } : {}) } } })
    onClose()
  }

  return (
    <Modal title="Create Player Character" onClose={onClose}>
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="field-label">Character Name</label>
          <input autoFocus value={name} onChange={e => setName(e.target.value)} className="input w-full" />
        </div>
        <div>
          <label className="field-label">Player Name <span className="text-slate-500">(optional)</span></label>
          <input value={ownerPlayerId} onChange={e => setOwnerPlayerId(e.target.value)} className="input w-full" />
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
