// REST API Entities

export interface Event {
  id: number;
  name: string;
  max_picks_per_team: number;
  max_teams_per_player: number;
  stipulations: Record<string, unknown>;
  status: 'pending' | 'in_progress' | 'completed';
  created_at: string;
  started_at: string | null;
  completed_at: string | null;
}

export interface Player {
  id: number;
  first_name: string;
  last_name: string;
  status: string;
  countryCode: string;
}

export interface User {
  id: number;
  eventID: number;
  username: string;
  created_at: string;
}

// Draft State

export interface Pick {
  userID: number;
  playerID: number;
  pickNumber: number;
  round: number;
  autoDraft: boolean;
}

// WebSocket Messages: Client -> Server

export interface StartDraftMessage {
  type: 'start_draft';
  eventID: number;
  pickOrder: number[];
  totalRounds: number;
  timerDuration: number;
  availablePlayers: number[];
}

export interface MakePickMessage {
  type: 'make_pick';
  userID: number;
  playerID: number;
}

export interface PauseDraftMessage {
  type: 'pause_draft';
}

export interface ResumeDraftMessage {
  type: 'resume_draft';
}

export type ClientMessage =
  | StartDraftMessage
  | MakePickMessage
  | PauseDraftMessage
  | ResumeDraftMessage;

// WebSocket Messages: Server -> Client

export interface DraftStartedMessage {
  type: 'draft_started';
  eventID: number;
  currentTurn: number;
  roundNumber: number;
  turnDeadline: number;
}

export interface PickMadeMessage {
  type: 'pick_made';
  userID: number;
  playerID: number;
  round: number;
  autoDraft: boolean;
}

export interface TurnChangedMessage {
  type: 'turn_changed';
  currentTurn: number;
  roundNumber: number;
  turnDeadline: number;
}

export interface DraftCompletedMessage {
  type: 'draft_completed';
  eventID: number;
  totalPicks: number;
  totalRounds: number;
}

export interface DraftPausedMessage {
  type: 'draft_paused';
  eventID: number;
  remainingTime: number;
}

export interface DraftResumedMessage {
  type: 'draft_resumed';
  eventID: number;
  currentTurn: number;
  roundNumber: number;
  turnDeadline: number;
}

export interface DraftStateMessage {
  type: 'draft_state';
  eventID: number;
  status: 'in_progress' | 'paused' | 'completed';
  currentTurn: number;
  roundNumber: number;
  currentPickIndex: number;
  totalRounds: number;
  pickOrder: number[];
  availablePlayers: number[];
  turnDeadline: number;
  remainingTime: number;
  pickHistory: Pick[];
}

export interface ErrorMessage {
  type: 'error';
  error: string;
}

export type ServerMessage =
  | DraftStartedMessage
  | PickMadeMessage
  | TurnChangedMessage
  | DraftCompletedMessage
  | DraftPausedMessage
  | DraftResumedMessage
  | DraftStateMessage
  | ErrorMessage;
