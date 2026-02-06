import { create } from "zustand";
import type { Player } from "../types";
import { getEventPlayers } from "../api/client";

interface PlayerList {
  eventPlayers: Player[];
  error: string | null;
  setEventPlayers: (eventID: number) => void;
}

const initialState = {
  eventPlayers: [],
  error: null,
}

export const usePlayerStore = create<PlayerList>((set) => ({
  ...initialState,

  setEventPlayers: (eventID) => {
    set({ error: null });
    getEventPlayers(eventID)
      .then((players) => set({ eventPlayers: players }))
      .catch((err) => set({ error: err.message }));
  }
}));