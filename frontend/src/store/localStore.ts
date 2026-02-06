import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface LocalState {
  eventID: number | null;
  setEventID: (eventID: number) => void;
  clear: () => void;
}

const initialState = {
  eventID: null,
};

export const useLocalStore = create<LocalState>()(
  persist(
    (set) => ({
      ...initialState,
      setEventID: (eventID) => set({ eventID }),
      clear: () => set(initialState),
    }),
    { name: 'draft-local-store' },
  ),
);
