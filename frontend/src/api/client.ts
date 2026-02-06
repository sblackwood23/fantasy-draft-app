import type { Event, Player, User } from '../types';

const API_BASE = 'http://localhost:8080';

interface FetchOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE';
  body?: unknown;
}

async function fetchJSON<T>(url: string, options: FetchOptions = {}): Promise<T> {
  const { method = 'GET', body } = options;

  const response = await fetch(`${API_BASE}${url}`, {
    method,
    headers: body ? { 'Content-Type': 'application/json' } : undefined,
    body: body ? JSON.stringify(body) : undefined,
  });

  if (!response.ok) {
    // Try to extract error message from JSON response
    const errorBody = await response.json().catch(() => null);
    const message = errorBody?.error || `HTTP ${response.status}: ${response.statusText}`;
    throw new Error(message);
  }
  return response.json();
}

export async function getEvents(): Promise<Event[]> {
  return fetchJSON<Event[]>('/events');
}

export async function getEvent(id: number): Promise<Event> {
  return fetchJSON<Event>(`/events/${id}`);
}

export async function getPlayers(): Promise<Player[]> {
  return fetchJSON<Player[]>('/players');
}

export async function getPlayer(id: number): Promise<Player> {
  return fetchJSON<Player>(`/players/${id}`);
}

export async function getUsers(): Promise<User[]> {
  return fetchJSON<User[]>('/users');
}

export async function getUser(id: number): Promise<User> {
  return fetchJSON<User>(`/users/${id}`);
}

export async function joinDraft(teamName: string, passkey: string): Promise<User> {
  return fetchJSON<User>(`/events/join`, {
    method: 'POST',
    body: { teamName, passkey },
  });
}

export async function getEventPlayers(eventID: number): Promise<Player[]> {
  return fetchJSON<Player[]>(`/events/${eventID}/players`);
}
