import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useGamesQuery, useCreateGameMutation, GamesDocument, GameSummary } from '../../../graphql/generated'
import { StatusBadge } from '../../../components/ui/StatusBadge'
import { Spinner } from '../../../components/ui/Spinner'
import { EmptyState } from '../../../components/ui/EmptyState'
import { ErrorMessage } from '../../../components/ui/ErrorMessage'
import { Modal } from '../../../components/ui/Modal'

export default function GamesPage() {
  const { data, loading, error } = useGamesQuery()
  const [showCreate, setShowCreate] = useState(false)

  if (loading) return <Spinner />
  if (error) return <div className="p-6"><ErrorMessage message={error.message} /></div>

  const games = data?.games ?? []

  return (
    <div className="p-6 max-w-5xl">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-slate-100">My Games</h1>
        <button
          onClick={() => setShowCreate(true)}
          className="btn-primary"
        >
          + New Game
        </button>
      </div>

      {games.length === 0 ? (
        <EmptyState
          title="Create your first Game"
          description="A Game is your campaign setting — the world, its story, and all the people in it."
          action={
            <button onClick={() => setShowCreate(true)} className="btn-primary">
              Create Game
            </button>
          }
        />
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {games.map(game => <GameCard key={game.id} game={game} />)}
        </div>
      )}

      {showCreate && <CreateGameModal onClose={() => setShowCreate(false)} />}
    </div>
  )
}

function GameCard({ game }: { game: GameSummary }) {
  const navigate = useNavigate()
  return (
    <button
      onClick={() => navigate(`/world-building/games/${game.id}`)}
      className="text-left bg-slate-800 border border-slate-700 rounded-lg p-4 hover:border-violet-600 hover:bg-slate-750 transition-colors group"
    >
      <div className="flex items-start justify-between gap-2 mb-2">
        <h2 className="font-semibold text-slate-100 group-hover:text-violet-300 transition-colors leading-tight">
          {game.name}
        </h2>
        <StatusBadge status={game.status} />
      </div>
      <p className="text-sm text-slate-400">
        {game.campaignCount} {game.campaignCount === 1 ? 'Campaign' : 'Campaigns'}
      </p>
      {game.lastActivityAt && (
        <p className="text-xs text-slate-500 mt-1">
          Last activity {formatRelativeTime(game.lastActivityAt)}
        </p>
      )}
    </button>
  )
}

function CreateGameModal({ onClose }: { onClose: () => void }) {
  const navigate = useNavigate()
  const [name, setName] = useState('')
  const [createGame, { loading, error }] = useCreateGameMutation({
    refetchQueries: [{ query: GamesDocument }],
  })

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!name.trim()) return
    const result = await createGame({ variables: { input: { name: name.trim() } } })
    const id = result.data?.createGame?.id
    if (id) {
      onClose()
      navigate(`/world-building/games/${id}`)
    }
  }

  return (
    <Modal title="Create New Game" onClose={onClose}>
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-slate-300 mb-1">
            Game Name
          </label>
          <input
            type="text"
            value={name}
            onChange={e => setName(e.target.value)}
            placeholder="Ashes & Chains"
            className="input w-full"
            autoFocus
          />
          <p className="mt-1 text-xs text-slate-500">The name for this campaign setting.</p>
        </div>

        {error && <ErrorMessage message={error.message} />}

        <div className="flex justify-end gap-3">
          <button type="button" onClick={onClose} className="btn-ghost">
            Cancel
          </button>
          <button type="submit" disabled={loading || !name.trim()} className="btn-primary">
            {loading ? 'Creating…' : 'Create Game'}
          </button>
        </div>
      </form>
    </Modal>
  )
}

function formatRelativeTime(iso: string): string {
  const ms = Date.now() - new Date(iso).getTime()
  const days = Math.floor(ms / 86_400_000)
  if (days === 0) return 'today'
  if (days === 1) return '1 day ago'
  if (days < 7) return `${days} days ago`
  const weeks = Math.floor(days / 7)
  if (weeks === 1) return '1 week ago'
  return `${weeks} weeks ago`
}
