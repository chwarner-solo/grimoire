import { create } from 'zustand'

interface WorldBuildingStore {
  activeGameId: string | null
  setActiveGameId: (id: string | null) => void
}

export const useWorldBuildingStore = create<WorldBuildingStore>((set) => ({
  activeGameId: null,
  setActiveGameId: (id) => set({ activeGameId: id }),
}))
