import { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import {
  useCampaignQuery, usePlayerCharactersQuery, useBeginCampaignFormationMutation,
  useAddCharacterToCampaignMutation, useCreatePlayerCharacterMutation,
  CampaignDocument, PlayerCharactersDocument, EntityStatus,
  PlayerCharacterSummary,
} from '../../../graphql/generated'
import { Spinner } from '../../../components/ui/Spinner'
import { ErrorMessage } from '../../../components/ui/ErrorMessage'

type Step = 1 | 2 | 3

export default function CampaignSetupPage() {
  const { gameId, campaignId } = useParams<{ gameId: string; campaignId: string }>()
  const navigate = useNavigate()

  const { data, loading, error } = useCampaignQuery({ variables: { id: campaignId!, gameId: gameId! } })
  const [beginFormation] = useBeginCampaignFormationMutation({
    refetchQueries: [{ query: CampaignDocument, variables: { id: campaignId!, gameId: gameId! } }],
  })

  if (loading) return <Spinner />
  if (error) return <div className="p-6"><ErrorMessage message={error.message} /></div>
  if (!data?.campaign) return <div className="p-6 text-slate-400">Campaign not found.</div>

  const campaign = data.campaign
  const step: Step =
    campaign.status === EntityStatus.New ? 1
    : campaign.status === EntityStatus.Forming ? 2
    : 3

  async function advanceToStep2() {
    await beginFormation({ variables: { input: { campaignId: campaignId!, gameId: gameId! } } })
  }

  return (
    <div className="p-6 max-w-2xl">
      <h1 className="text-sm text-slate-400 mb-1">Setting up: {campaign.name}</h1>

      <StepIndicator current={step} />

      <div className="mt-8">
        {step === 1 && <Step1 campaignName={campaign.name} onAdvance={advanceToStep2} />}
        {step === 2 && (
          <Step2
            gameId={gameId!}
            campaignId={campaignId!}
            characters={campaign.characters}
          />
        )}
        {step === 3 && (
          <Step3
            characterCount={campaign.characters.length}
            onGoToOverview={() => navigate(`/world-building/games/${gameId}/campaigns`)}
            onStartSession={() => navigate('/session')}
          />
        )}
      </div>
    </div>
  )
}

function StepIndicator({ current }: { current: Step }) {
  const steps = ['Campaign', 'Party', 'Ready']
  return (
    <div className="flex items-center gap-0 mt-4">
      {steps.map((label, i) => {
        const stepNum = (i + 1) as Step
        const isDone = stepNum < current
        const isCurrent = stepNum === current
        return (
          <div key={label} className="flex items-center">
            <div className="flex flex-col items-center">
              <div className={`w-3 h-3 rounded-full border-2 transition-colors ${
                isDone ? 'bg-violet-500 border-violet-500' :
                isCurrent ? 'bg-violet-400 border-violet-400' :
                'bg-transparent border-slate-600'
              }`} />
              <span className={`text-xs mt-1 ${isCurrent ? 'text-violet-300' : 'text-slate-500'}`}>
                {label}
              </span>
            </div>
            {i < steps.length - 1 && (
              <div className={`w-24 h-0.5 mb-4 ${isDone ? 'bg-violet-500' : 'bg-slate-700'}`} />
            )}
          </div>
        )
      })}
    </div>
  )
}

function Step1({ campaignName, onAdvance }: { campaignName: string; onAdvance: () => void }) {
  return (
    <div className="space-y-6">
      <div>
        <label className="field-label">Campaign Name</label>
        <div className="input w-full text-slate-400 cursor-default">{campaignName}</div>
      </div>
      <div className="flex justify-end">
        <button onClick={onAdvance} className="btn-primary">Begin Party Formation ▶</button>
      </div>
    </div>
  )
}

function Step2({ gameId, campaignId, characters }: {
  gameId: string; campaignId: string; characters: PlayerCharacterSummary[]
}) {
  const navigate = useNavigate()
  const [showAddExisting, setShowAddExisting] = useState(false)
  const [showCreateNew, setShowCreateNew] = useState(false)
  const [newCharName, setNewCharName] = useState('')

  const { data: pcsData } = usePlayerCharactersQuery({ variables: { gameId } })
  const [addChar] = useAddCharacterToCampaignMutation({
    refetchQueries: [{ query: CampaignDocument, variables: { id: campaignId, gameId } }],
  })
  const [createPC] = useCreatePlayerCharacterMutation({
    refetchQueries: [{ query: PlayerCharactersDocument, variables: { gameId } }],
  })

  const campaignCharIds = new Set(characters.map(c => c.id))
  const available = (pcsData?.playerCharacters ?? []).filter(pc =>
    !campaignCharIds.has(pc.id) && pc.status === EntityStatus.Active
  )

  async function handleCreateAndAdd() {
    if (!newCharName.trim()) return
    const result = await createPC({ variables: { input: { gameId, name: newCharName.trim() } } })
    const id = result.data?.createPlayerCharacter?.id
    if (id) {
      await addChar({ variables: { input: { campaignId, gameId, characterId: id } } })
      setNewCharName('')
      setShowCreateNew(false)
    }
  }

  const canContinue = characters.length >= 1

  return (
    <div className="space-y-4">
      <h2 className="font-medium text-slate-200">Add Player Characters</h2>

      <div className="bg-slate-800 border border-slate-700 rounded-lg divide-y divide-slate-700">
        {characters.map(c => (
          <div key={c.id} className="flex items-center justify-between px-4 py-3">
            <span className="text-slate-200">{c.name}</span>
          </div>
        ))}
        {characters.length === 0 && (
          <div className="px-4 py-3 text-sm text-slate-500">No characters yet.</div>
        )}
      </div>

      <div className="flex gap-3">
        {available.length > 0 && (
          <div className="relative">
            {!showAddExisting ? (
              <button onClick={() => setShowAddExisting(true)} className="btn-secondary text-sm">
                + Add Existing Character
              </button>
            ) : (
              <select
                autoFocus
                className="input text-sm"
                defaultValue=""
                onChange={e => {
                  addChar({ variables: { input: { campaignId, gameId, characterId: e.target.value } } })
                  setShowAddExisting(false)
                }}
                onBlur={() => setShowAddExisting(false)}
              >
                <option value="" disabled>Select character…</option>
                {available.map(pc => <option key={pc.id} value={pc.id}>{pc.name}</option>)}
              </select>
            )}
          </div>
        )}

        {!showCreateNew ? (
          <button onClick={() => setShowCreateNew(true)} className="btn-secondary text-sm">
            + Create New
          </button>
        ) : (
          <div className="flex items-center gap-2">
            <input
              autoFocus
              value={newCharName}
              onChange={e => setNewCharName(e.target.value)}
              placeholder="Character name"
              className="input text-sm"
              onKeyDown={e => e.key === 'Enter' && handleCreateAndAdd()}
            />
            <button onClick={handleCreateAndAdd} disabled={!newCharName.trim()} className="text-xs text-violet-400">✓</button>
            <button onClick={() => setShowCreateNew(false)} className="text-xs text-slate-500">✕</button>
          </div>
        )}
      </div>

      {!canContinue && (
        <p className="text-sm text-slate-500">Need at least 1 character to start playing.</p>
      )}

      <div className="flex justify-between pt-2">
        <button
          onClick={() => navigate(-1)}
          className="btn-ghost"
        >
          ← Back
        </button>
        <button
          disabled={!canContinue}
          className="btn-primary"
          onClick={() => {}}
        >
          Continue to Review ▶
        </button>
      </div>
    </div>
  )
}

function Step3({ characterCount, onGoToOverview, onStartSession }: {
  characterCount: number
  onGoToOverview: () => void; onStartSession: () => void
}) {
  return (
    <div className="space-y-6">
      <ul className="space-y-2">
        {[
          'Campaign created',
          `Party assembled (${characterCount} character${characterCount !== 1 ? 's' : ''})`,
          'Ready for Session 1',
        ].map(item => (
          <li key={item} className="flex items-center gap-2 text-slate-300">
            <span className="text-emerald-400">✓</span>
            {item}
          </li>
        ))}
      </ul>

      <p className="text-sm text-slate-400">
        When you're ready to play, switch to Session mode to start the first session.
      </p>

      <div className="flex gap-3">
        <button onClick={onGoToOverview} className="btn-secondary">Go to Campaign Overview</button>
        <button onClick={onStartSession} className="btn-primary">Start Session Now ▶</button>
      </div>
    </div>
  )
}
