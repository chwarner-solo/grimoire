import { useState } from 'react'
import { useParams } from 'react-router-dom'
import {
  useMacGuffinsQuery, useCreateMacGuffinMutation,
  MacGuffinsDocument, MacGuffinSummary,
} from '../../../graphql/generated'
import { Spinner } from '../../../components/ui/Spinner'
import { EmptyState } from '../../../components/ui/EmptyState'
import { ErrorMessage } from '../../../components/ui/ErrorMessage'
import { Modal } from '../../../components/ui/Modal'

export default function MacGuffinsPage() {
  const { gameId } = useParams<{ gameId: string }>()
  const { data, loading, error } = useMacGuffinsQuery({ variables: { gameId: gameId! } })
  const [showCreate, setShowCreate] = useState(false)
  const [selectedId, setSelectedId] = useState<string | null>(null)

  if (loading) return <Spinner />
  if (error) return <div className="p-6"><ErrorMessage message={error.message} /></div>

  const macguffins = data?.macguffins ?? []

  return (
    <div className="p-6 max-w-3xl">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-xl font-bold text-slate-100">MacGuffins</h1>
        <button onClick={() => setShowCreate(true)} className="btn-primary">+ Add MacGuffin</button>
      </div>

      {macguffins.length === 0 ? (
        <EmptyState
          title="No MacGuffins yet"
          description="MacGuffins are narratively significant items — artifacts, documents, relics — that drive the story."
          action={<button onClick={() => setShowCreate(true)} className="btn-primary">Create MacGuffin</button>}
        />
      ) : (
        <div className="bg-slate-800 border border-slate-700 rounded-lg divide-y divide-slate-700">
          {macguffins.map(m => (
            <button
              key={m.id}
              onClick={() => setSelectedId(m.id)}
              className={`w-full flex items-center justify-between px-4 py-3 hover:bg-slate-700/50 transition-colors text-left
                ${selectedId === m.id ? 'bg-violet-900/20' : ''}`}
            >
              <span className="font-medium text-slate-200">{m.name}</span>
              <span className="text-sm text-slate-400">{describePossessor(m)}</span>
            </button>
          ))}
        </div>
      )}

      {showCreate && <CreateMacGuffinModal gameId={gameId!} onClose={() => setShowCreate(false)} />}
    </div>
  )
}

function describePossessor(m: MacGuffinSummary): string {
  if (!m.possessor) return 'Possessed: None'
  const p = m.possessor
  if (p.__typename === 'NPCRef') return `Possessed: ${p.name}`
  if (p.__typename === 'PlayerCharacterRef') return `Carried by: ${p.name}`
  if (p.__typename === 'LocationRef') return `At: ${p.name}`
  return 'Unknown possessor'
}

function CreateMacGuffinModal({ gameId, onClose }: { gameId: string; onClose: () => void }) {
  const [name, setName] = useState('')
  const [createMacGuffin, { loading, error }] = useCreateMacGuffinMutation({
    refetchQueries: [{ query: MacGuffinsDocument, variables: { gameId } }],
  })

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!name.trim()) return
    await createMacGuffin({ variables: { input: { gameId, name: name.trim() } } })
    onClose()
  }

  return (
    <Modal title="Create MacGuffin" onClose={onClose}>
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
