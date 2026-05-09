# Flamingo Chat

A Go-based anonymous chat server with matchmaking, real-time WebSocket transport, and a clean modular architecture.

## Project Structure

```
flamingo_chat/
│   go.mod
│   go.sum
│
├───cmd/
│   └───server/
│       main.go              ← entry point
│
└───internal/
    ├───app/
    │   handler.go           ← event routing, wires server ↔ transport
    │
    ├───chat/
    │   chat.go              ← Chat struct
    │   manager.go           ← in-memory chat store
    │   message.go           ← Message struct
    │
    ├───events/
    │   builder.go           ← BuildEvent (serialize)
    │   events.go            ← all event types and payload structs
    │   parser.go            ← ParseEvent (deserialize)
    │
    ├───matchmaking/
    │   matcher.go           ← FindMatch logic
    │   queue.go             ← Queue and QueueEntry
    │
    ├───server/
    │   presence.go          ← GetPresenceStats
    │   server.go            ← core orchestrator, all business logic
    │   state.go             ← in-memory state (users, sessions)
    │   worker.go            ← background session cleanup
    │
    ├───transport/
    │   └───websocket/
    │       client.go        ← one Client per WebSocket connection
    │       hub.go           ← registry of all connected clients
    │
    └───user/
        session.go           ← Session struct
        user.go              ← User struct and status constants
```

## Architecture

The project is split into four distinct layers. Each layer only talks to the layer next to it.

```
[Browser]
    ↕ WebSocket (raw JSON bytes)
[transport/websocket]    — dumb pipe, just reads and writes bytes
    ↕ parsed events
[app]                    — routes events to the right server method
    ↕ method calls + callbacks
[server]                 — owns all state, all business logic
    ↕
[chat / matchmaking / user]   — pure domain types and logic
```

**Why this separation matters:**
- The transport layer knows nothing about matchmaking or chat logic
- The server knows nothing about WebSockets
- The app layer is the only place they meet
- `main.go` stays clean — it only creates things and starts them

## Running the server

```bash
go run ./cmd/server/main.go
```

Server starts on `:8080`. WebSocket endpoint: `ws://localhost:8080/ws`

## Running the tests

```bash
# all tests
go test ./...

# specific packages with output
go test ./internal/matchmaking/... -v
go test ./internal/server/... -v
go test ./internal/integration/... -v

# one specific test
go test ./internal/server/... -run TestLeaveChat_NotifiesPartner -v
```

## Connecting a client manually

```bash
# install wscat
npm install -g wscat

# connect
wscat -c ws://localhost:8080/ws

# first message must always be init
{"type":"init","payload":{"user_id":""}}

# join the queue
{"type":"join_queue","payload":{"gender":"male","preference":"female"}}

# send a message (after being matched)
{"type":"send_message","payload":{"chat_id":"CHAT_ID_HERE","content":"hello"}}
```

## User lifecycle

```
connect → init → ready → join_queue → queue_joined → match_found → [chat] → leave_chat
                                                                          ↑
                                                              chat_ended (partner left)
```

## Event protocol summary

All communication is JSON events over WebSocket with the shape `{"type": "...", "payload": {...}}`.

**Client sends:** `init`, `join_queue`, `send_message`, `ping`, `leave_chat`

**Server sends:** `ready`, `queue_joined`, `match_found`, `message_received`, `chat_ended`, `error`

See [FRONTEND_GUIDE.md](./FRONTEND_GUIDE.md) for the complete frontend integration reference including all payload schemas, the full JavaScript client example, and UI state design notes.

## Key design decisions

- **No REST API** — everything over WebSocket, one persistent connection per user
- **No authentication** — users are identified by a short random hex ID stored on the client
- **No message history** — chats are ephemeral, messages are not persisted between sessions
- **Mutex on server, channels on hub** — server state is protected by a single mutex; hub uses Go channels so its client map is only ever touched by one goroutine
- **Callbacks not interfaces** — the server exposes `OnMatchFound`, `OnMessage`, `OnChatEnded` function fields; the app layer sets these during `Wire()`, keeping `main.go` to ~10 lines