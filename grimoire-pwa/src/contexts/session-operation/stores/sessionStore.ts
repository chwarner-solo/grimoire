import { create } from 'zustand'

interface SessionStore {
  sessionId: string | null
  campaignId: string | null
  setSession: (sessionId: string, campaignId: string) => void
  clearSession: () => void
}

export const useSessionStore = create<SessionStore>((set) => ({
  sessionId: null,
  campaignId: null,
  setSession: (sessionId, campaignId) => set({ sessionId, campaignId }),
  clearSession: () => set({ sessionId: null, campaignId: null }),
}))
