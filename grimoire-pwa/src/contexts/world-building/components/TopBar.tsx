import { useNavigate } from 'react-router-dom'

export function TopBar({ gameName }: { gameName?: string }) {
  const navigate = useNavigate()

  return (
    <header className="h-14 bg-slate-900 border-b border-slate-700 flex items-center px-4 shrink-0 gap-4">
      <button
        onClick={() => navigate('/world-building')}
        className="text-lg font-bold text-violet-400 tracking-wide hover:text-violet-300 transition-colors"
      >
        GRIMOIRE
      </button>

      <div className="flex items-center gap-1 bg-slate-800 border border-slate-700 rounded px-3 py-1 text-sm">
        <span className="text-slate-300 font-medium">World Building</span>
        <span className="text-slate-500">▼</span>
      </div>

      {gameName && (
        <span className="text-sm text-slate-400 truncate max-w-xs">{gameName}</span>
      )}

      <div className="ml-auto">
        <div className="w-8 h-8 rounded-full bg-slate-700 border border-slate-600 flex items-center justify-center text-xs text-slate-300">
          GM
        </div>
      </div>
    </header>
  )
}
