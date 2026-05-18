import { useState, useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import {
  useNpcQuery, useUpdateNpcContentMutation, useActivateNpcMutation, useArchiveNpcMutation,
  NpcDocument, EntityStatus,
} from '../../../graphql/generated'
import { StatusBadge } from '../../../components/ui/StatusBadge'
import { Spinner } from '../../../components/ui/Spinner'
import { ErrorMessage } from '../../../components/ui/ErrorMessage'

export default function NpcDetailPage() {
  const { gameId, npcId } = useParams<{ gameId: string; npcId: string }>()
  const navigate = useNavigate()
  const refetchVars = { query: NpcDocument, variables: { id: npcId!, gameId: gameId! } }

  const { data, loading, error } = useNpcQuery({ variables: { id: npcId!, gameId: gameId! } })
  const [updateContent, { loading: saving, error: saveError }] = useUpdateNpcContentMutation({ refetchQueries: [refetchVars] })
  const [activateNPC] = useActivateNpcMutation({ refetchQueries: [refetchVars] })
  const [archiveNPC] = useArchiveNpcMutation({ refetchQueries: [refetchVars] })

  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [playerDesc, setPlayerDesc] = useState('')
  const [dirty, setDirty] = useState(false)

  const npc = data?.npc

  useEffect(() => {
    if (npc && !dirty) {
      setName(npc.name)
      setDescription(npc.description)
      setPlayerDesc(npc.playerDescription)
    }
  }, [npc, dirty])

  if (loading) return <Spinner />
  if (error) return <div className="p-6"><ErrorMessage message={error.message} /></div>
  if (!npc) return <div className="p-6 text-slate-400">NPC not found.</div>

  const isDraft = npc.status === EntityStatus.Draft

  async function handleSave() {
    await updateContent({ variables: { input: { npcId: npcId!, gameId: gameId!, name, description, playerDesc } } })
    setDirty(false)
  }

  return (
    <div className="p-6 max-w-2xl space-y-6">
      <div className="flex items-center gap-2 text-sm text-slate-500">
        <button onClick={() => navigate(`/world-building/games/${gameId}/characters`)} className="hover:text-violet-400">
          ← Characters
        </button>
      </div>

      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <h1 className="text-xl font-bold text-slate-100">{npc.name}</h1>
          <StatusBadge status={npc.status} />
        </div>

        <div className="flex items-center gap-2">
          {npc.status === EntityStatus.Active && (
            <button
              onClick={() => {
                if (confirm(`Archive "${npc.name}"? MacGuffins will drop to their last known location.`)) {
                  archiveNPC({ variables: { input: { npcId: npcId!, gameId: gameId! } } })
                    .then(() => navigate(`/world-building/games/${gameId}/characters`))
                }
              }}
              className="text-xs text-slate-500 hover:text-red-400"
            >
              Archive
            </button>
          )}
        </div>
      </div>

      <div className="space-y-4">
        <div>
          <label className="field-label">Name</label>
          <input
            value={name}
            onChange={e => { setName(e.target.value); setDirty(true) }}
            className="input w-full"
          />
        </div>

        <div>
          <label className="field-label">Description <span className="text-slate-500">(GM-only)</span></label>
          <textarea
            value={description}
            onChange={e => { setDescription(e.target.value); setDirty(true) }}
            className="input w-full h-28 resize-none"
          />
        </div>

        <div>
          <label className="field-label">Player Description <span className="text-slate-500">(shown on reveal)</span></label>
          <textarea
            value={playerDesc}
            onChange={e => { setPlayerDesc(e.target.value); setDirty(true) }}
            className="input w-full h-28 resize-none"
          />
        </div>
      </div>

      {npc.factionMemberships.length > 0 && (
        <div>
          <h2 className="text-sm font-medium text-slate-300 mb-2">Faction Memberships</h2>
          <div className="space-y-1">
            {npc.factionMemberships.map(m => (
              <div key={m.faction.id} className="text-sm text-slate-400">
                {m.faction.name} — {m.rank}
              </div>
            ))}
          </div>
          <p className="text-xs text-slate-600 mt-1">Membership is managed on the Faction screen.</p>
        </div>
      )}

      {saveError && <ErrorMessage message={saveError.message} />}

      <div className="flex gap-3">
        <button onClick={handleSave} disabled={!dirty || saving} className="btn-primary flex-1">
          {saving ? 'Saving…' : 'Save Changes'}
        </button>

        {isDraft && (
          <button
            onClick={() => activateNPC({ variables: { input: { npcId: npcId!, gameId: gameId! } } })}
            className="btn-secondary"
          >
            Activate ▶
          </button>
        )}
      </div>
    </div>
  )
}
