import { NavLink, useParams } from 'react-router-dom'

const NAV_ITEMS = [
  { label: 'Overview',   path: '' },
  { label: 'Narrative',  path: '/narrative' },
  { label: 'Locations',  path: '/locations' },
  { label: 'Factions',   path: '/factions' },
  { label: 'Characters', path: '/characters' },
  { label: 'MacGuffins', path: '/macguffins' },
  { label: 'Campaigns',  path: '/campaigns' },
]

function navClass({ isActive }: { isActive: boolean }) {
  return `block px-3 py-2 rounded text-sm transition-colors ${
    isActive
      ? 'bg-violet-900/50 text-violet-300 font-medium'
      : 'text-slate-400 hover:text-slate-200 hover:bg-slate-700/50'
  }`
}

export function Sidebar({ gameName }: { gameName?: string }) {
  const { gameId } = useParams<{ gameId: string }>()

  return (
    <aside className="w-60 bg-slate-800 border-r border-slate-700 flex flex-col shrink-0">
      {gameId ? (
        <>
          <div className="px-4 py-3 border-b border-slate-700">
            <p className="text-sm font-semibold text-slate-200 truncate">{gameName ?? '…'}</p>
          </div>
          <nav className="flex-1 overflow-y-auto px-2 py-3 space-y-0.5">
            {NAV_ITEMS.map(({ label, path }) => (
              <NavLink
                key={label}
                to={`/world-building/games/${gameId}${path}`}
                end={path === ''}
                className={navClass}
              >
                {label}
              </NavLink>
            ))}
          </nav>
        </>
      ) : (
        <div className="px-4 py-3">
          <p className="text-xs text-slate-500 uppercase tracking-wide font-medium">Your Games</p>
        </div>
      )}
    </aside>
  )
}
