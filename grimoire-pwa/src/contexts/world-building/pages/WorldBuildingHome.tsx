import { Routes, Route, Navigate, useParams } from 'react-router-dom'
import { useGameQuery } from '../../../graphql/generated'
import { AppShell } from '../components/AppShell'
import GamesPage from './GamesPage'
import GameOverviewPage from './GameOverviewPage'
import LocationsPage from './LocationsPage'
import NarrativePage from './NarrativePage'
import FactionsPage from './FactionsPage'
import FactionDetailPage from './FactionDetailPage'
import CharactersPage from './CharactersPage'
import NpcDetailPage from './NpcDetailPage'
import MacGuffinsPage from './MacGuffinsPage'
import CampaignsPage from './CampaignsPage'
import CampaignSetupPage from './CampaignSetupPage'

export default function WorldBuildingHome() {
  return (
    <Routes>
      <Route element={<AppShell />}>
        <Route index element={<GamesPage />} />
      </Route>
      <Route path="games/:gameId/*" element={<GameShell />} />
    </Routes>
  )
}

function GameShell() {
  const { gameId } = useParams<{ gameId: string }>()
  const { data } = useGameQuery({ variables: { id: gameId! } })
  const gameName = data?.game?.name

  return (
    <AppShell gameName={gameName}>
      <Routes>
        <Route index element={<GameOverviewPage />} />
        <Route path="locations" element={<LocationsPage />} />
        <Route path="narrative" element={<NarrativePage />} />
        <Route path="factions" element={<FactionsPage />} />
        <Route path="factions/:factionId" element={<FactionDetailPage />} />
        <Route path="characters" element={<CharactersPage />} />
        <Route path="characters/npcs/:npcId" element={<NpcDetailPage />} />
        <Route path="macguffins" element={<MacGuffinsPage />} />
        <Route path="campaigns" element={<CampaignsPage />} />
        <Route path="campaigns/:campaignId/setup" element={<CampaignSetupPage />} />
        <Route path="*" element={<Navigate to="" replace />} />
      </Routes>
    </AppShell>
  )
}
