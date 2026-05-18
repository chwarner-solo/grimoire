import { Outlet } from 'react-router-dom'
import { TopBar } from './TopBar'
import { Sidebar } from './Sidebar'

interface Props {
  gameName?: string
  children?: React.ReactNode
}

export function AppShell({ gameName, children }: Props) {
  return (
    <div className="flex flex-col h-screen bg-slate-900 text-slate-100">
      <TopBar gameName={gameName} />
      <div className="flex flex-1 overflow-hidden">
        <Sidebar gameName={gameName} />
        <main className="flex-1 overflow-y-auto">
          {children ?? <Outlet />}
        </main>
      </div>
    </div>
  )
}
