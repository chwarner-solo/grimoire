import { useNavigate, useParams } from 'react-router-dom'
import { useGameQuery, EntityStatus, EntityCountSummary, CampaignSummary } from '../../../graphql/generated'
import { StatusBadge } from '../../../components/ui/StatusBadge'
import { Spinner } from '../../../components/ui/Spinner'
import { ErrorMessage } from '../../../components/ui/ErrorMessage'

export default function GameOverviewPage() {
  const { gameId } = useParams<{ gameId: string }>()
  const { data, loading, error } = useGameQuery({ variables: { id: gameId! } })

  if (loading) return <Spinner />
  if (error) return <div className="p-6"><ErrorMessage message={error.message} /></div>
  if (!data?.game) return <div className="p-6 text-slate-400">Game not found.</div>

  const { game } = data

  return (
    <div className="p-6 max-w-4xl space-y-6">
      <div className="flex items-center gap-3">
        <h1 className="text-2xl font-bold text-slate-100">{game.name}</h1>
        <StatusBadge status={game.status} />
      </div>

      <NextStepCard game={game} gameId={gameId!} />

      <section>
        <h2 className="text-sm font-medium text-slate-400 uppercase tracking-wide mb-3">
          World Summary
        </h2>
        <div className="grid grid-cols-3 gap-4">
          <WorldCountCard title="Locations" counts={game.locationSummary} />
          <WorldCountCard title="Factions"  counts={game.factionSummary} />
          <WorldCountCard title="NPCs"      counts={game.npcSummary} />
        </div>
      </section>

      <CampaignsSection gameId={gameId!} campaigns={game.campaigns} />
    </div>
  )
}

type GameData = NonNullable<ReturnType<typeof useGameQuery>['data']>['game']

function NextStepCard({ game, gameId }: { game: NonNullable<GameData>; gameId: string }) {
  const navigate = useNavigate()
  const message = deriveNextStep(game, gameId)

  if (!message) return null

  return (
    <div className="bg-slate-800 border border-violet-800/50 rounded-lg p-4">
      <p className="text-sm font-medium text-slate-300 mb-1">Next Step</p>
      <p className="text-slate-400 text-sm mb-3">{message.text}</p>
      <button
        onClick={() => navigate(message.href)}
        className="btn-secondary text-sm"
      >
        {message.cta}
      </button>
    </div>
  )
}

function deriveNextStep(game: NonNullable<GameData>, gameId: string): { text: string; cta: string; href: string } | null {
  const base = `/world-building/games/${gameId}`

  if (game.status === EntityStatus.New) {
    return { text: 'Author your first story beat to activate the Master Narrative.', cta: 'Go to Narrative', href: `${base}/narrative` }
  }
  if (game.status === EntityStatus.Draft) {
    const hasLocations = game.locationSummary.active + game.locationSummary.draft > 0
    if (!hasLocations) return { text: 'Build your first location.', cta: 'Go to Locations', href: `${base}/locations` }
    if (game.campaigns.length === 0) return { text: 'Create a Campaign to start playing.', cta: 'Create Campaign', href: `${base}/campaigns` }
    return { text: `Set up your party in ${game.campaigns[0].name}.`, cta: 'Set Up Party', href: `${base}/campaigns/${game.campaigns[0].id}/setup` }
  }
  if (game.status === EntityStatus.Active) {
    return { text: 'A session is in progress. Switch to Session mode to continue.', cta: 'Go to Session', href: '/session' }
  }
  if (game.status === EntityStatus.Idle) {
    return { text: 'Your game is ready for the next session.', cta: 'Start Session', href: '/session' }
  }
  return null
}

function WorldCountCard({ title, counts }: { title: string; counts: EntityCountSummary }) {
  return (
    <div className="bg-slate-800 border border-slate-700 rounded-lg p-4">
      <p className="text-sm font-medium text-slate-300 mb-2">{title}</p>
      <dl className="space-y-1 text-sm">
        {counts.draft > 0 && (
          <div className="flex justify-between text-slate-400">
            <dt>Draft</dt><dd>{counts.draft}</dd>
          </div>
        )}
        {counts.active > 0 && (
          <div className="flex justify-between text-emerald-400">
            <dt>Active</dt><dd>{counts.active}</dd>
          </div>
        )}
        {counts.idle > 0 && (
          <div className="flex justify-between text-amber-400">
            <dt>Idle</dt><dd>{counts.idle}</dd>
          </div>
        )}
        {counts.archived > 0 && (
          <div className="flex justify-between text-slate-500">
            <dt>Archived</dt><dd>{counts.archived}</dd>
          </div>
        )}
        {counts.draft + counts.active + counts.idle + counts.archived === 0 && (
          <p className="text-slate-600">None yet</p>
        )}
      </dl>
    </div>
  )
}

function CampaignsSection({ gameId, campaigns }: { gameId: string; campaigns: CampaignSummary[] }) {
  const navigate = useNavigate()
  return (
    <section>
      <h2 className="text-sm font-medium text-slate-400 uppercase tracking-wide mb-3">Campaigns</h2>
      {campaigns.length === 0 ? (
        <div className="bg-slate-800 border border-slate-700 rounded-lg p-4 flex items-center justify-between">
          <p className="text-slate-400 text-sm">No campaigns yet.</p>
          <button
            onClick={() => navigate(`/world-building/games/${gameId}/campaigns`)}
            className="btn-secondary text-sm"
          >
            Create Campaign
          </button>
        </div>
      ) : (
        <div className="space-y-2">
          {campaigns.map(c => (
            <button
              key={c.id}
              onClick={() => navigate(`/world-building/games/${gameId}/campaigns`)}
              className="w-full text-left bg-slate-800 border border-slate-700 rounded-lg px-4 py-3 hover:border-violet-600 transition-colors flex items-center justify-between"
            >
              <span className="font-medium text-slate-200">{c.name}</span>
              <StatusBadge status={c.status} />
            </button>
          ))}
        </div>
      )}
    </section>
  )
}
