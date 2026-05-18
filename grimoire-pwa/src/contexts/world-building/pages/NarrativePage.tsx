import { useState } from 'react'
import { useParams } from 'react-router-dom'
import {
  useMasterNarrativeQuery, useBeatQuery, useCreateMasterBeatMutation,
  useUpdateBeatContentMutation, useAddBeatPrerequisiteMutation,
  MasterNarrativeDocument, BeatDocument,
  BeatSummary, BeatType,
} from '../../../graphql/generated'
import { Spinner } from '../../../components/ui/Spinner'
import { EmptyState } from '../../../components/ui/EmptyState'
import { ErrorMessage } from '../../../components/ui/ErrorMessage'

const BEAT_TYPE_LABELS: Record<BeatType, string> = {
  [BeatType.Required]:         'Required',
  [BeatType.Optional]:         'Optional',
  [BeatType.CampaignSpecific]: 'Campaign-Specific',
}

const BEAT_TYPE_STYLE: Record<BeatType, string> = {
  [BeatType.Required]:         'border-violet-600 text-violet-300',
  [BeatType.Optional]:         'border-slate-500 text-slate-300',
  [BeatType.CampaignSpecific]: 'border-dashed border-slate-500 text-slate-400',
}

export default function NarrativePage() {
  const { gameId } = useParams<{ gameId: string }>()
  const [view, setView] = useState<'dag' | 'list'>('dag')
  const [selectedBeatId, setSelectedBeatId] = useState<string | null>(null)
  const [showCreate, setShowCreate] = useState(false)
  const { data, loading, error } = useMasterNarrativeQuery({ variables: { gameId: gameId! } })

  if (loading) return <Spinner />
  if (error) return <div className="p-6"><ErrorMessage message={error.message} /></div>

  const beats = data?.masterNarrative?.beats ?? []
  const narrativeId = data?.masterNarrative?.id ?? ''

  return (
    <div className="flex h-full">
      <div className="flex-1 flex flex-col overflow-hidden">
        <div className="flex items-center justify-between px-6 py-4 border-b border-slate-700">
          <h1 className="font-semibold text-slate-100">Narrative</h1>
          <div className="flex items-center gap-2">
            <div className="flex border border-slate-600 rounded overflow-hidden text-xs">
              <button
                onClick={() => setView('dag')}
                className={`px-3 py-1.5 ${view === 'dag' ? 'bg-violet-900 text-violet-300' : 'text-slate-400 hover:bg-slate-700'}`}
              >
                DAG View
              </button>
              <button
                onClick={() => setView('list')}
                className={`px-3 py-1.5 ${view === 'list' ? 'bg-violet-900 text-violet-300' : 'text-slate-400 hover:bg-slate-700'}`}
              >
                List View
              </button>
            </div>
            <button onClick={() => setShowCreate(true)} className="btn-primary text-sm">
              + Beat
            </button>
          </div>
        </div>

        <div className="flex-1 overflow-y-auto p-6">
          {beats.length === 0 ? (
            <EmptyState
              title="No beats yet"
              description="Story beats form the backbone of your narrative. Create the first one."
              action={<button onClick={() => setShowCreate(true)} className="btn-primary">Create Beat</button>}
            />
          ) : view === 'list' ? (
            <BeatListView beats={beats} selectedId={selectedBeatId} onSelect={setSelectedBeatId} />
          ) : (
            <BeatDagView beats={beats} selectedId={selectedBeatId} onSelect={setSelectedBeatId} />
          )}
        </div>
      </div>

      {selectedBeatId && (
        <BeatDetailPanel
          beatId={selectedBeatId}
          gameId={gameId!}
          allBeats={beats}
          onClose={() => setSelectedBeatId(null)}
        />
      )}

      {showCreate && (
        <CreateBeatModal gameId={gameId!} narrativeId={narrativeId} onClose={() => setShowCreate(false)} />
      )}
    </div>
  )
}

function BeatListView({ beats, selectedId, onSelect }: {
  beats: BeatSummary[]; selectedId: string | null; onSelect: (id: string) => void
}) {
  return (
    <div className="bg-slate-800 border border-slate-700 rounded-lg divide-y divide-slate-700">
      <div className="grid grid-cols-12 gap-4 px-4 py-2 text-xs font-medium text-slate-500 uppercase tracking-wide">
        <div className="col-span-5">Name</div>
        <div className="col-span-3">Type</div>
        <div className="col-span-2">Prerequisites</div>
      </div>
      {beats.map(beat => (
        <button
          key={beat.id}
          onClick={() => onSelect(beat.id)}
          className={`w-full grid grid-cols-12 gap-4 px-4 py-3 text-left hover:bg-slate-700/50 transition-colors
            ${selectedId === beat.id ? 'bg-violet-900/20' : ''}`}
        >
          <div className="col-span-5 text-sm font-medium text-slate-200">{beat.name}</div>
          <div className="col-span-3 text-xs text-slate-400">{BEAT_TYPE_LABELS[beat.beatType]}</div>
          <div className="col-span-2 text-xs text-slate-500">{beat.prerequisiteCount}</div>
        </button>
      ))}
    </div>
  )
}

function BeatDagView({ beats, selectedId, onSelect }: {
  beats: BeatSummary[]; selectedId: string | null; onSelect: (id: string) => void
}) {
  return (
    <div className="flex flex-wrap gap-4">
      {beats.map(beat => (
        <button
          key={beat.id}
          onClick={() => onSelect(beat.id)}
          className={`border rounded-lg px-4 py-3 text-left min-w-36 max-w-48 transition-colors
            ${BEAT_TYPE_STYLE[beat.beatType]}
            ${selectedId === beat.id ? 'bg-violet-900/30' : 'bg-slate-800 hover:bg-slate-700/50'}`}
        >
          <p className="text-xs text-slate-500 mb-1">[{BEAT_TYPE_LABELS[beat.beatType]}]</p>
          <p className="text-sm font-medium leading-tight">{beat.name}</p>
          {beat.prerequisiteCount > 0 && (
            <p className="text-xs text-slate-500 mt-1">↳ {beat.prerequisiteCount} prereqs</p>
          )}
        </button>
      ))}
    </div>
  )
}

function BeatDetailPanel({ beatId, gameId, allBeats, onClose }: {
  beatId: string; gameId: string; allBeats: BeatSummary[]
  onClose: () => void
}) {
  const { data, loading } = useBeatQuery({ variables: { id: beatId, gameId } })
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [playerDesc, setPlayerDesc] = useState('')
  const [dirty, setDirty] = useState(false)

  const [updateBeat, { loading: saving }] = useUpdateBeatContentMutation({
    refetchQueries: [{ query: BeatDocument, variables: { id: beatId, gameId } }],
  })
  const [addPrereq] = useAddBeatPrerequisiteMutation({
    refetchQueries: [{ query: BeatDocument, variables: { id: beatId, gameId } }],
  })

  const beat = data?.beat
  const initialized = !!beat && name === ''

  if (initialized) {
    setName(beat.name)
    setDescription(beat.description)
    setPlayerDesc(beat.playerDescription)
  }

  if (loading) return (
    <aside className="w-96 border-l border-slate-700 flex items-center justify-center">
      <Spinner />
    </aside>
  )

  if (!beat) return null

  async function handleSave() {
    await updateBeat({ variables: { input: { beatId, gameId, name, description, playerDesc } } })
    setDirty(false)
  }

  const prereqIds = new Set(beat.prerequisites.map(p => p.id))
  const availableForPrereq = allBeats.filter(b => b.id !== beatId && !prereqIds.has(b.id))

  return (
    <aside className="w-96 border-l border-slate-700 flex flex-col overflow-hidden">
      <div className="flex items-center justify-between px-4 py-3 border-b border-slate-700">
        <span className="text-sm text-slate-400">{BEAT_TYPE_LABELS[beat.beatType]}</span>
        <button onClick={onClose} className="text-slate-400 hover:text-slate-200">✕</button>
      </div>

      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        <div>
          <label className="field-label">Name</label>
          <input
            value={name}
            onChange={e => { setName(e.target.value); setDirty(true) }}
            className="input w-full"
          />
        </div>

        <div>
          <label className="field-label">GM Description</label>
          <textarea
            value={description}
            onChange={e => { setDescription(e.target.value); setDirty(true) }}
            className="input w-full h-24 resize-none"
          />
        </div>

        <div>
          <label className="field-label">Player Description</label>
          <textarea
            value={playerDesc}
            onChange={e => { setPlayerDesc(e.target.value); setDirty(true) }}
            className="input w-full h-24 resize-none"
          />
        </div>

        <div>
          <div className="flex items-center justify-between mb-2">
            <label className="field-label mb-0">Prerequisites</label>
            {availableForPrereq.length > 0 && (
              <PrereqPicker beats={availableForPrereq} onAdd={prereqId =>
                addPrereq({ variables: { input: { beatId, prerequisiteId: prereqId, gameId } } })
              } />
            )}
          </div>
          <div className="bg-slate-800 border border-slate-700 rounded divide-y divide-slate-700">
            {beat.prerequisites.map(p => (
              <div key={p.id} className="px-3 py-2 text-sm text-slate-300">{p.name}</div>
            ))}
            {beat.prerequisites.length === 0 && (
              <div className="px-3 py-2 text-sm text-slate-500">None</div>
            )}
          </div>
        </div>
      </div>

      <div className="px-4 py-3 border-t border-slate-700">
        <button
          onClick={handleSave}
          disabled={!dirty || saving}
          className="btn-primary w-full"
        >
          {saving ? 'Saving…' : 'Save Changes'}
        </button>
      </div>
    </aside>
  )
}

function PrereqPicker({ beats, onAdd }: { beats: BeatSummary[]; onAdd: (id: string) => void }) {
  const [open, setOpen] = useState(false)

  if (!open) return (
    <button onClick={() => setOpen(true)} className="text-xs text-violet-400 hover:text-violet-300">
      + Add prereq
    </button>
  )

  return (
    <select
      autoFocus
      className="input text-xs"
      defaultValue=""
      onChange={e => { onAdd(e.target.value); setOpen(false) }}
      onBlur={() => setOpen(false)}
    >
      <option value="" disabled>Select beat…</option>
      {beats.map(b => <option key={b.id} value={b.id}>{b.name}</option>)}
    </select>
  )
}

function CreateBeatModal({ gameId, narrativeId: _, onClose }: {
  gameId: string; narrativeId: string; onClose: () => void
}) {
  const [name, setName] = useState('')
  const [beatType, setBeatType] = useState<BeatType>(BeatType.Required)
  const [createBeat, { loading, error }] = useCreateMasterBeatMutation({
    refetchQueries: [{ query: MasterNarrativeDocument, variables: { gameId } }],
  })

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!name.trim()) return
    await createBeat({ variables: { input: { gameId, name: name.trim(), beatType } } })
    onClose()
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60" onClick={onClose}>
      <div className="bg-slate-800 border border-slate-700 rounded-lg p-6 w-full max-w-sm" onClick={e => e.stopPropagation()}>
        <h2 className="text-lg font-semibold mb-4">New Beat</h2>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="field-label">Name</label>
            <input autoFocus value={name} onChange={e => setName(e.target.value)} className="input w-full" />
          </div>
          <div>
            <label className="field-label">Type</label>
            <select value={beatType} onChange={e => setBeatType(e.target.value as BeatType)} className="input w-full">
              {Object.entries(BEAT_TYPE_LABELS).map(([val, label]) => (
                <option key={val} value={val}>{label}</option>
              ))}
            </select>
          </div>
          {error && <ErrorMessage message={error.message} />}
          <div className="flex justify-end gap-3">
            <button type="button" onClick={onClose} className="btn-ghost">Cancel</button>
            <button type="submit" disabled={loading || !name.trim()} className="btn-primary">Create</button>
          </div>
        </form>
      </div>
    </div>
  )
}
