import { EntityStatus } from '../../graphql/generated'

const statusConfig: Record<EntityStatus, { label: string; className: string }> = {
  [EntityStatus.New]:      { label: 'New',      className: 'bg-slate-700 text-slate-300' },
  [EntityStatus.Draft]:    { label: 'Draft',    className: 'bg-slate-600 text-slate-200' },
  [EntityStatus.Forming]:  { label: 'Forming',  className: 'bg-blue-900 text-blue-300' },
  [EntityStatus.Active]:   { label: 'Active',   className: 'bg-emerald-900 text-emerald-300' },
  [EntityStatus.Idle]:     { label: 'Idle',     className: 'bg-amber-900 text-amber-300' },
  [EntityStatus.Complete]: { label: 'Complete', className: 'bg-violet-900 text-violet-300' },
  [EntityStatus.Archived]: { label: 'Archived', className: 'bg-slate-800 text-slate-500' },
  [EntityStatus.Retired]:  { label: 'Retired',  className: 'bg-slate-800 text-slate-500' },
}

interface Props {
  status: EntityStatus
}

export function StatusBadge({ status }: Props) {
  const config = statusConfig[status]
  return (
    <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${config.className}`}>
      {config.label}
    </span>
  )
}
