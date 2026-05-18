import { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import {
  useFactionQuery, useNpcsQuery, useActivateFactionMutation, useAddFactionMemberMutation,
  useAddStandingLevelMutation, useDeclareAllyMutation, useDeclareWarMutation,
  useArchiveFactionMutation, useFactionsQuery,
  FactionDocument, EntityStatus, FactionMember, StandingLevel, FactionRef,
} from '../../../graphql/generated'
import { StatusBadge } from '../../../components/ui/StatusBadge'
import { Spinner } from '../../../components/ui/Spinner'
import { ErrorMessage } from '../../../components/ui/ErrorMessage'

export default function FactionDetailPage() {
  const { gameId, factionId } = useParams<{ gameId: string; factionId: string }>()
  const navigate = useNavigate()
  const refetchVars = { query: FactionDocument, variables: { id: factionId!, gameId: gameId! } }

  const { data, loading, error } = useFactionQuery({ variables: { id: factionId!, gameId: gameId! } })
  const [activateFaction] = useActivateFactionMutation({ refetchQueries: [refetchVars] })
  const [archiveFaction] = useArchiveFactionMutation({ refetchQueries: [refetchVars] })

  if (loading) return <Spinner />
  if (error) return <div className="p-6"><ErrorMessage message={error.message} /></div>
  if (!data?.faction) return <div className="p-6 text-slate-400">Faction not found.</div>

  const faction = data.faction
  const isDraft = faction.status === EntityStatus.Draft

  return (
    <div className="p-6 max-w-2xl space-y-6">
      <div className="flex items-center gap-2 text-sm text-slate-500">
        <button onClick={() => navigate(`/world-building/games/${gameId}/factions`)} className="hover:text-violet-400">
          ← Factions
        </button>
      </div>

      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <h1 className="text-xl font-bold text-slate-100">{faction.name}</h1>
          <StatusBadge status={faction.status} />
        </div>
        <button
          onClick={() => {
            if (confirm(`Archive "${faction.name}"? This is terminal.`)) {
              archiveFaction({ variables: { input: { factionId: factionId!, gameId: gameId! } } })
                .then(() => navigate(`/world-building/games/${gameId}/factions`))
            }
          }}
          className="text-xs text-slate-500 hover:text-red-400"
        >
          ⋯
        </button>
      </div>

      <MembersSection factionId={factionId!} gameId={gameId!} members={faction.members} />
      <StandingLevelsSection factionId={factionId!} gameId={gameId!} levels={faction.standingLevels} />
      <RelationshipsSection factionId={factionId!} gameId={gameId!} allies={faction.allies} enemies={faction.enemies} />

      {isDraft && (
        <button
          onClick={() => activateFaction({ variables: { input: { factionId: factionId!, gameId: gameId! } } })}
          className="btn-primary w-full"
        >
          Activate Faction ▶
        </button>
      )}
    </div>
  )
}

function MembersSection({ factionId, gameId, members }: {
  factionId: string; gameId: string; members: FactionMember[]
}) {
  const [adding, setAdding] = useState(false)
  const [npcId, setNpcId] = useState('')
  const [rank, setRank] = useState('')
  const { data: npcsData } = useNpcsQuery({ variables: { gameId, excludeArchived: true } })
  const [addMember, { loading }] = useAddFactionMemberMutation({
    refetchQueries: [{ query: FactionDocument, variables: { id: factionId, gameId } }],
  })

  const npcs = npcsData?.npcs ?? []
  const memberNpcIds = new Set(members.map(m => m.npc.id))
  const available = npcs.filter(n => !memberNpcIds.has(n.id))

  async function handleAdd(e: React.FormEvent) {
    e.preventDefault()
    if (!npcId || !rank.trim()) return
    await addMember({ variables: { input: { factionId, gameId, npcId, rank: rank.trim() } } })
    setNpcId('')
    setRank('')
    setAdding(false)
  }

  return (
    <section>
      <div className="flex items-center justify-between mb-2">
        <h2 className="text-sm font-medium text-slate-300">Members</h2>
        {available.length > 0 && (
          <button onClick={() => setAdding(true)} className="text-xs text-violet-400 hover:text-violet-300">
            + Add Member
          </button>
        )}
      </div>
      <div className="bg-slate-800 border border-slate-700 rounded-lg divide-y divide-slate-700">
        {members.map(m => (
          <div key={m.npc.id} className="px-4 py-2 flex items-center justify-between text-sm">
            <span className="text-slate-200">{m.npc.name}</span>
            <span className="text-slate-400">{m.rank}</span>
          </div>
        ))}
        {adding && (
          <form onSubmit={handleAdd} className="px-4 py-2 flex items-center gap-2">
            <select value={npcId} onChange={e => setNpcId(e.target.value)} className="input text-sm flex-1">
              <option value="" disabled>Select NPC…</option>
              {available.map(n => <option key={n.id} value={n.id}>{n.name}</option>)}
            </select>
            <input value={rank} onChange={e => setRank(e.target.value)} placeholder="Rank" className="input text-sm w-32" />
            <button type="submit" disabled={loading} className="text-xs text-violet-400">✓</button>
            <button type="button" onClick={() => setAdding(false)} className="text-xs text-slate-500">✕</button>
          </form>
        )}
        {members.length === 0 && !adding && (
          <div className="px-4 py-3 text-sm text-slate-500">No members yet.</div>
        )}
      </div>
    </section>
  )
}

function StandingLevelsSection({ factionId, gameId, levels }: {
  factionId: string; gameId: string; levels: StandingLevel[]
}) {
  const [adding, setAdding] = useState(false)
  const [levelName, setLevelName] = useState('')
  const [threshold, setThreshold] = useState('')
  const [addLevel, { loading }] = useAddStandingLevelMutation({
    refetchQueries: [{ query: FactionDocument, variables: { id: factionId, gameId } }],
  })

  async function handleAdd(e: React.FormEvent) {
    e.preventDefault()
    const t = parseInt(threshold, 10)
    if (!levelName.trim() || isNaN(t)) return
    await addLevel({ variables: { input: { factionId, gameId, name: levelName.trim(), threshold: t, ordinal: levels.length + 1 } } })
    setLevelName('')
    setThreshold('')
    setAdding(false)
  }

  return (
    <section>
      <div className="flex items-center justify-between mb-2">
        <h2 className="text-sm font-medium text-slate-300">Standing Levels</h2>
        <button onClick={() => setAdding(true)} className="text-xs text-violet-400 hover:text-violet-300">
          + Add Level
        </button>
      </div>
      <div className="bg-slate-800 border border-slate-700 rounded-lg divide-y divide-slate-700">
        <div className="grid grid-cols-3 px-4 py-2 text-xs text-slate-500 font-medium uppercase tracking-wide">
          <span>Ordinal</span><span>Name</span><span>Threshold</span>
        </div>
        {levels.map(l => (
          <div key={l.ordinal} className="grid grid-cols-3 px-4 py-2 text-sm text-slate-300">
            <span>{l.ordinal}</span><span>{l.name}</span><span>{l.threshold}</span>
          </div>
        ))}
        {adding && (
          <form onSubmit={handleAdd} className="flex items-center gap-2 px-4 py-2">
            <input value={levelName} onChange={e => setLevelName(e.target.value)} placeholder="Name" className="input text-sm flex-1" />
            <input value={threshold} onChange={e => setThreshold(e.target.value)} placeholder="Threshold" type="number" className="input text-sm w-24" />
            <button type="submit" disabled={loading} className="text-xs text-violet-400">✓</button>
            <button type="button" onClick={() => setAdding(false)} className="text-xs text-slate-500">✕</button>
          </form>
        )}
        {levels.length === 0 && !adding && (
          <div className="px-4 py-3 text-sm text-slate-500">No standing levels defined.</div>
        )}
      </div>
    </section>
  )
}

function RelationshipsSection({ factionId, gameId, allies, enemies }: {
  factionId: string; gameId: string; allies: FactionRef[]; enemies: FactionRef[]
}) {
  const { data } = useFactionsQuery({ variables: { gameId } })
  const [declareAlly] = useDeclareAllyMutation({
    refetchQueries: [{ query: FactionDocument, variables: { id: factionId, gameId } }],
  })
  const [declareWar] = useDeclareWarMutation({
    refetchQueries: [{ query: FactionDocument, variables: { id: factionId, gameId } }],
  })

  const allFactions = data?.factions ?? []
  const relatedIds = new Set([factionId, ...allies.map(a => a.id), ...enemies.map(e => e.id)])
  const available = allFactions.filter(f => !relatedIds.has(f.id))

  return (
    <section>
      <h2 className="text-sm font-medium text-slate-300 mb-2">Relationships</h2>
      <div className="bg-slate-800 border border-slate-700 rounded-lg divide-y divide-slate-700 p-4 space-y-3">
        <RelationshipRow
          label="Allied with"
          items={allies}
          available={available}
          onAdd={id => declareAlly({ variables: { input: { factionId, targetId: id, gameId } } })}
          addLabel="+ Ally"
        />
        <RelationshipRow
          label="At war with"
          items={enemies}
          available={available}
          onAdd={id => declareWar({ variables: { input: { factionId, targetId: id, gameId } } })}
          addLabel="+ Enemy"
        />
      </div>
    </section>
  )
}

function RelationshipRow({ label, items, available, onAdd, addLabel }: {
  label: string; items: FactionRef[]
  available: { id: string; name: string }[]
  onAdd: (id: string) => void; addLabel: string
}) {
  const [open, setOpen] = useState(false)
  return (
    <div className="flex items-start gap-3 text-sm">
      <span className="text-slate-500 w-24 shrink-0">{label}:</span>
      <div className="flex flex-wrap gap-1 flex-1">
        {items.map(i => <span key={i.id} className="text-slate-300">{i.name}</span>)}
        {items.length === 0 && <span className="text-slate-600">—</span>}
      </div>
      {available.length > 0 && !open && (
        <button onClick={() => setOpen(true)} className="text-xs text-violet-400 hover:text-violet-300 shrink-0">
          {addLabel}
        </button>
      )}
      {open && (
        <select
          autoFocus
          className="input text-xs"
          defaultValue=""
          onChange={e => { onAdd(e.target.value); setOpen(false) }}
          onBlur={() => setOpen(false)}
        >
          <option value="" disabled>Select…</option>
          {available.map(f => <option key={f.id} value={f.id}>{f.name}</option>)}
        </select>
      )}
    </div>
  )
}
