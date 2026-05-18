import { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import {
  useCampaignsQuery, useCreateCampaignMutation,
  CampaignsDocument, CampaignSummary,
} from '../../../graphql/generated'
import { StatusBadge } from '../../../components/ui/StatusBadge'
import { Spinner } from '../../../components/ui/Spinner'
import { EmptyState } from '../../../components/ui/EmptyState'
import { ErrorMessage } from '../../../components/ui/ErrorMessage'
import { Modal } from '../../../components/ui/Modal'

export default function CampaignsPage() {
  const { gameId } = useParams<{ gameId: string }>()
  const { data, loading, error } = useCampaignsQuery({ variables: { gameId: gameId! } })
  const [showCreate, setShowCreate] = useState(false)

  if (loading) return <Spinner />
  if (error) return <div className="p-6"><ErrorMessage message={error.message} /></div>

  const campaigns = data?.campaigns ?? []

  return (
    <div className="p-6 max-w-3xl">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-xl font-bold text-slate-100">Campaigns</h1>
        <button onClick={() => setShowCreate(true)} className="btn-primary">+ New Campaign</button>
      </div>

      {campaigns.length === 0 ? (
        <EmptyState
          title="No campaigns yet"
          description="A Campaign is a table playing through this Game's story."
          action={<button onClick={() => setShowCreate(true)} className="btn-primary">Create Campaign</button>}
        />
      ) : (
        <div className="space-y-3">
          {campaigns.map(c => <CampaignCard key={c.id} campaign={c} gameId={gameId!} />)}
        </div>
      )}

      {showCreate && <CreateCampaignModal gameId={gameId!} onClose={() => setShowCreate(false)} />}
    </div>
  )
}

function CampaignCard({ campaign, gameId }: { campaign: CampaignSummary; gameId: string }) {
  const navigate = useNavigate()
  return (
    <div className="bg-slate-800 border border-slate-700 rounded-lg p-4 flex items-center justify-between">
      <div>
        <h2 className="font-semibold text-slate-200">{campaign.name}</h2>
        <div className="flex items-center gap-4 mt-1 text-sm text-slate-400">
          <span>{campaign.characterCount} character{campaign.characterCount !== 1 ? 's' : ''}</span>
          <span>{campaign.sessionCount} session{campaign.sessionCount !== 1 ? 's' : ''}</span>
        </div>
      </div>
      <div className="flex items-center gap-3">
        <StatusBadge status={campaign.status} />
        <button
          onClick={() => navigate(`/world-building/games/${gameId}/campaigns/${campaign.id}/setup`)}
          className="text-violet-400 hover:text-violet-300 text-sm"
        >
          →
        </button>
      </div>
    </div>
  )
}

function CreateCampaignModal({ gameId, onClose }: { gameId: string; onClose: () => void }) {
  const navigate = useNavigate()
  const [name, setName] = useState('')
  const [createCampaign, { loading, error }] = useCreateCampaignMutation({
    refetchQueries: [{ query: CampaignsDocument, variables: { gameId } }],
  })

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!name.trim()) return
    const result = await createCampaign({ variables: { input: { gameId, name: name.trim() } } })
    const id = result.data?.createCampaign?.id
    if (id) {
      onClose()
      navigate(`/world-building/games/${gameId}/campaigns/${id}/setup`)
    }
  }

  return (
    <Modal title="New Campaign" onClose={onClose}>
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="field-label">Campaign Name</label>
          <input autoFocus value={name} onChange={e => setName(e.target.value)} placeholder="Table A — Friday Night" className="input w-full" />
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
