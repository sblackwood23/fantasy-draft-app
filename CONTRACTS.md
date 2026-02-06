# API Contracts

This document defines the contracts between the frontend client and backend server.

## HTTP REST Endpoints

Base URL: `http://localhost:8080`

### Events

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/events` | List all events |
| GET | `/events/{id}` | Get a single event |
| POST | `/events` | Create a new event |
| PUT | `/events/{id}` | Update an event |
| DELETE | `/events/{id}` | Delete an event |

**Event Object:**
```json
{
  "id": 1,
  "name": "2024 Fantasy Draft",
  "max_picks_per_team": 5,
  "max_teams_per_player": 1,
  "stipulations": {},
  "status": "pending",
  "passkey": "secret123",
  "created_at": "2024-01-01T00:00:00Z",
  "started_at": null,
  "completed_at": null
}
```

### Players

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/players` | List all players |
| GET | `/players/{id}` | Get a single player |
| POST | `/players` | Create a new player |
| PUT | `/players/{id}` | Update a player |
| DELETE | `/players/{id}` | Delete a player |

**Player Object:**
```json
{
  "id": 1,
  "first_name": "John",
  "last_name": "Doe",
  "status": "active",
  "country": "USA"
}
```

### Users

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/users` | List all users |
| GET | `/users/{id}` | Get a single user |
| POST | `/users` | Create a new user |
| PUT | `/users/{id}` | Update a user |
| DELETE | `/users/{id}` | Delete a user |

**User Object:**
```json
{
  "id": 1,
  "event_id": 1,
  "username": "team_alpha",
  "created_at": "2024-01-01T00:00:00Z"
}
```

### Draft Room

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/events/{id}/join` | Join/authenticate for a draft room |
| POST | `/events/{id}/draft-room` | Create a draft room for an event |
| GET | `/events/{id}/draft-room` | Get draft room state |

#### `POST /events/{id}/join`

Validates passkey and registers/authenticates a user for the draft. Used when entering a draft room.

**Request:**
```json
{
  "team_name": "Team Alpha",
  "passkey": "secret123"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `team_name` | string | Yes | The team/username for this draft |
| `passkey` | string | Yes | The event's passkey for authentication |

**Response (201 Created):** New user registered
```json
{
  "id": 1,
  "event_id": 1,
  "username": "Team Alpha",
  "created_at": "2024-01-01T00:00:00Z"
}
```

**Response (200 OK):** Existing user (reconnection)
```json
{
  "id": 1,
  "event_id": 1,
  "username": "Team Alpha",
  "created_at": "2024-01-01T00:00:00Z"
}
```

**Error Responses:**

| Status | Error | Description |
|--------|-------|-------------|
| 400 | `team_name is required` | Missing team_name in request |
| 401 | `invalid passkey` | Passkey doesn't match event's passkey |
| 404 | `event not found` | Event ID doesn't exist |
| 409 | `draft room is full` | Event already has 12 teams and username doesn't match existing user |

### Health Check

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Check server health |

---

## WebSocket Connection

**Endpoint:** `ws://localhost:8080/ws/draft`

All messages are JSON objects with a `type` field indicating the message type.

---

## WebSocket Messages: Client to Server

### `start_draft`

Starts a new draft. Should be sent by an admin user.

```json
{
  "type": "start_draft",
  "eventID": 1,
  "pickOrder": [1, 2, 3, 4],
  "totalRounds": 5,
  "timerDuration": 60,
  "availablePlayers": [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
}
```

| Field | Type | Description |
|-------|------|-------------|
| `eventID` | number | ID of the event to start |
| `pickOrder` | number[] | Array of user IDs in draft order |
| `totalRounds` | number | Number of rounds in the draft |
| `timerDuration` | number | Seconds each user has to make a pick |
| `availablePlayers` | number[] | Array of player IDs available to draft |

### `make_pick`

Makes a pick during the draft.

```json
{
  "type": "make_pick",
  "userID": 1,
  "playerID": 5
}
```

| Field | Type | Description |
|-------|------|-------------|
| `userID` | number | ID of the user making the pick |
| `playerID` | number | ID of the player being drafted |

### `pause_draft`

Pauses an in-progress draft.

```json
{
  "type": "pause_draft"
}
```

### `resume_draft`

Resumes a paused draft.

```json
{
  "type": "resume_draft"
}
```

---

## WebSocket Messages: Server to Client

### `draft_started`

Broadcast when a draft begins.

```json
{
  "type": "draft_started",
  "eventID": 1,
  "currentTurn": 1,
  "roundNumber": 1,
  "turnDeadline": 1704067260
}
```

| Field | Type | Description |
|-------|------|-------------|
| `eventID` | number | ID of the event |
| `currentTurn` | number | User ID whose turn it is |
| `roundNumber` | number | Current round number |
| `turnDeadline` | number | Unix timestamp when the turn expires |

### `pick_made`

Broadcast when a pick is made (manually or via auto-draft).

```json
{
  "type": "pick_made",
  "userID": 1,
  "playerID": 5,
  "round": 1,
  "autoDraft": false
}
```

| Field | Type | Description |
|-------|------|-------------|
| `userID` | number | ID of the user who made the pick |
| `playerID` | number | ID of the player drafted |
| `round` | number | Round in which the pick was made |
| `autoDraft` | boolean | `true` if pick was auto-drafted due to timer expiry |

### `turn_changed`

Broadcast when the turn advances to the next user.

```json
{
  "type": "turn_changed",
  "currentTurn": 2,
  "roundNumber": 1,
  "turnDeadline": 1704067320
}
```

| Field | Type | Description |
|-------|------|-------------|
| `currentTurn` | number | User ID whose turn it is now |
| `roundNumber` | number | Current round number |
| `turnDeadline` | number | Unix timestamp when the turn expires |

### `draft_completed`

Broadcast when the draft finishes.

```json
{
  "type": "draft_completed",
  "eventID": 1,
  "totalPicks": 20,
  "totalRounds": 5
}
```

| Field | Type | Description |
|-------|------|-------------|
| `eventID` | number | ID of the completed event |
| `totalPicks` | number | Total number of picks made |
| `totalRounds` | number | Total rounds in the draft |

### `draft_paused`

Broadcast when a draft is paused.

```json
{
  "type": "draft_paused",
  "eventID": 1,
  "remainingTime": 45.5
}
```

| Field | Type | Description |
|-------|------|-------------|
| `eventID` | number | ID of the event |
| `remainingTime` | number | Seconds remaining on the turn timer when paused |

### `draft_resumed`

Broadcast when a paused draft resumes.

```json
{
  "type": "draft_resumed",
  "eventID": 1,
  "currentTurn": 2,
  "roundNumber": 1,
  "turnDeadline": 1704067320
}
```

| Field | Type | Description |
|-------|------|-------------|
| `eventID` | number | ID of the event |
| `currentTurn` | number | User ID whose turn it is |
| `roundNumber` | number | Current round number |
| `turnDeadline` | number | Unix timestamp when the turn expires |

### `draft_state`

Sent to newly connected clients during an active draft (for reconnection sync).

```json
{
  "type": "draft_state",
  "eventID": 1,
  "status": "in_progress",
  "currentTurn": 2,
  "roundNumber": 1,
  "currentPickIndex": 3,
  "totalRounds": 5,
  "pickOrder": [1, 2, 3, 4],
  "availablePlayers": [5, 6, 7, 8, 9, 10],
  "turnDeadline": 1704067320,
  "remainingTime": 0,
  "pickHistory": [
    {"userID": 1, "playerID": 1, "pickNumber": 1, "round": 1, "autoDraft": false},
    {"userID": 2, "playerID": 2, "pickNumber": 2, "round": 1, "autoDraft": false},
    {"userID": 3, "playerID": 3, "pickNumber": 3, "round": 1, "autoDraft": true}
  ]
}
```

| Field | Type | Description |
|-------|------|-------------|
| `eventID` | number | ID of the event |
| `status` | string | Draft status: `in_progress`, `paused`, or `completed` |
| `currentTurn` | number | User ID whose turn it is |
| `roundNumber` | number | Current round number |
| `currentPickIndex` | number | Current position in pick sequence (0-indexed) |
| `totalRounds` | number | Total rounds in the draft |
| `pickOrder` | number[] | Array of user IDs in draft order |
| `availablePlayers` | number[] | Array of player IDs still available |
| `turnDeadline` | number | Unix timestamp when the turn expires |
| `remainingTime` | number | Seconds remaining (used when paused) |
| `pickHistory` | object[] | Array of all picks made so far |

### `error`

Sent to a single client when an error occurs.

```json
{
  "type": "error",
  "error": "not your turn"
}
```

| Field | Type | Description |
|-------|------|-------------|
| `error` | string | Error message describing what went wrong |

---

## Draft Flow

1. Clients connect to `/ws/draft`
2. **If draft already in progress:** Server sends `draft_state` to the connecting client
3. Admin sends `start_draft` with configuration
4. Server broadcasts `draft_started` to all clients
5. Current user sends `make_pick` before timer expires
6. Server broadcasts `pick_made` and `turn_changed`
7. If timer expires, server auto-drafts and broadcasts `pick_made` with `autoDraft: true`
8. Optionally, admin can send `pause_draft` / `resume_draft` to control the draft
9. Repeat until all rounds complete
10. Server broadcasts `draft_completed`

## Reconnection

Clients connecting mid-draft automatically receive the full draft state via `draft_state` message. This includes:
- Current turn and round information
- Timer deadline (or remaining time if paused)
- List of available players
- Complete pick history for rebuilding the draft board

## Snake Draft Order

The draft uses snake ordering:
- Round 1: User 1 -> 2 -> 3 -> 4
- Round 2: User 4 -> 3 -> 2 -> 1
- Round 3: User 1 -> 2 -> 3 -> 4
- (pattern continues...)
