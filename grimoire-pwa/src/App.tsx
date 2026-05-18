import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import WorldBuildingContext from './contexts/world-building/pages/WorldBuildingHome'
import SessionOperationContext from './contexts/session-operation/pages/SessionHome'

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Navigate to="/world-building" replace />} />
        <Route path="/world-building/*" element={<WorldBuildingContext />} />
        <Route path="/session/*" element={<SessionOperationContext />} />
      </Routes>
    </BrowserRouter>
  )
}
