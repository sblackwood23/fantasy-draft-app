import { create } from 'zustand';
import type { Pick, ServerMessage } from '../types';

type ConnectionStatus = 'disconnected' | 'connecting' | 'connected';
type DraftStatus = 'idle' | 'in_progress' | 'paused' | 'completed';

interface DraftState {
  // Connection
  connectionStatus: ConnectionStatus;

  // Draft state
  draftStatus: DraftStatus;
  eventID: number | null;
  currentTurn: number | null;
  roundNumber: number;
  totalRounds: number;
  currentPickIndex: number;
  pickOrder: number[];
  availablePlayerIDs: number[];
  pickHistory: Pick[];
  turnDeadline: number | null;
  remainingTime: number;

  // Error
  lastError: string | null;

  // Actions
  setConnectionStatus: (status: ConnectionStatus) => void;
  setEventID: (eventID: number) => void;
  handleServerMessage: (message: ServerMessage) => void;
  reset: () => void;
}

const initialState = {
  connectionStatus: 'disconnected' as ConnectionStatus,
  draftStatus: 'idle' as DraftStatus,
  eventID: null,
  currentTurn: null,
  roundNumber: 0,
  totalRounds: 0,
  currentPickIndex: 0,
  pickOrder: [],
  availablePlayerIDs: [],
  pickHistory: [],
  turnDeadline: null,
  remainingTime: 0,
  lastError: null,
};

export const useDraftStore = create<DraftState>((set) => ({
  ...initialState,

  setConnectionStatus: (status) => set({ connectionStatus: status }),
  setEventID: (eventID) => set({ eventID }),

  handleServerMessage: (message) => {
    switch (message.type) {
      case 'draft_started':
        set({
          draftStatus: 'in_progress',
          eventID: message.eventID,
          currentTurn: message.currentTurn,
          roundNumber: message.roundNumber,
          turnDeadline: message.turnDeadline,
          lastError: null,
        });
        break;

      case 'draft_state':
        set({
          draftStatus: message.status === 'in_progress' ? 'in_progress'
                     : message.status === 'paused' ? 'paused'
                     : 'completed',
          eventID: message.eventID,
          currentTurn: message.currentTurn,
          roundNumber: message.roundNumber,
          totalRounds: message.totalRounds,
          currentPickIndex: message.currentPickIndex,
          pickOrder: message.pickOrder,
          availablePlayerIDs: message.availablePlayers,
          pickHistory: message.pickHistory,
          turnDeadline: message.turnDeadline,
          remainingTime: message.remainingTime,
          lastError: null,
        });
        break;

      case 'pick_made':
        set((state) => ({
          pickHistory: [
            ...state.pickHistory,
            {
              userID: message.userID,
              playerID: message.playerID,
              pickNumber: state.pickHistory.length + 1,
              round: message.round,
              autoDraft: message.autoDraft,
            },
          ],
          availablePlayerIDs: state.availablePlayerIDs.filter(
            (id) => id !== message.playerID
          ),
        }));
        break;

      case 'turn_changed':
        set({
          currentTurn: message.currentTurn,
          roundNumber: message.roundNumber,
          turnDeadline: message.turnDeadline,
        });
        break;

      case 'draft_completed':
        set({
          draftStatus: 'completed',
          eventID: message.eventID,
          totalRounds: message.totalRounds,
        });
        break;

      case 'draft_paused':
        set({
          draftStatus: 'paused',
          remainingTime: message.remainingTime,
        });
        break;

      case 'draft_resumed':
        set({
          draftStatus: 'in_progress',
          currentTurn: message.currentTurn,
          roundNumber: message.roundNumber,
          turnDeadline: message.turnDeadline,
        });
        break;

      case 'error':
        set({ lastError: message.error });
        break;
    }
  },

  reset: () => set(initialState),
}));
