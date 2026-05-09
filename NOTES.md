# Flamingo Chat

A simple Go-based chat server with matchmaking, real-time WebSocket transport, and clean modular design.

## What this project includes

- Matchmaking system for pairing users in private chats
- WebSocket transport layer for real-time messaging
- Event-based protocol for chat actions
- User session handling and presence tracking
- Background cleanup of inactive sessions

## Project Structure

```
│   go.mod
│   go.sum
│
├───cmd
│   └───server
│       main.go
│
└───internal
    ├───app
    │
    ├───chat
    │   chat.go
    │   manager.go
    │   message.go
    │
    ├───events
    │   builder.go
    │   events.go
    │   parser.go
    │
    ├───matchmaking
    │   matcher.go
    │   queue.go
    │
    ├───server
    │   presence.go
    │   server.go
    │   state.go
    │   worker.go
    │
    ├───transport
    │   └───websocket
    │
    ├───user
    │   session.go
    │   user.go
    │
    └───utils
```

## Core components

### Entry point
- `cmd/server/main.go`
  - Starts the chat server

### Chat management
- `internal/chat/chat.go`
  - `GetOtherUserID`
- `internal/chat/manager.go`
  - `NewManager`
  - `CreateChat`
  - `EndChat`
  - `SendMessage`
  - `GetChat`
  - `RemoveChat`
- `internal/chat/message.go`
  - Message definitions and helper logic

### Event handling
- `internal/events/builder.go`
  - `BuildEvent`
- `internal/events/events.go`
  - Event definitions and payload types
- `internal/events/parser.go`
  - `ParseEvent`

### Matchmaking
- `internal/matchmaking/matcher.go`
  - `FindMatch`
- `internal/matchmaking/queue.go`
  - `NewQueue`
  - `AddUser`
  - `RemoveUser`
  - `RemoveUsers`

### Server logic and state
- `internal/server/presence.go`
  - `GetPresenceStats`
- `internal/server/server.go`
  - `NewServer`
  - `CreateUser`
  - `SetPreferences`
  - `SetOnline`
  - `JoinQueue`
  - `TryMatch`
  - `SendMessage`
  - `CreateSession`
  - `PingSession`
  - `DisconnectSession`
  - `LeaveChat`
- `internal/server/state.go`
  - Server state management
- `internal/server/worker.go`
  - `StartCleanupWorker`
  - `cleanupInactiveSessions`

### WebSocket transport
- `internal/transport/websocket/`

### User management
- `internal/user/session.go`
  - Session handling and expiry
- `internal/user/user.go`
  - User profile and preferences

## Code summary

A quick list of packages and exported symbols:

- `flamingo_chat`
  - `main`
- `flamingo_chat/internal/events`
  - `BuildEvent`
  - `Event`
  - `JoinQueuePayload`
  - `SendMessagePayload`
  - `MatchFoundPayload`
  - `MessageReceivedPayload`
  - `ErrorPayload`
  - `NextPartnerPayload`
  - `ChatEndedPayload`
  - `ParseEvent`
- `flamingo_chat/internal/chat`
  - `Chat`
  - `GetOtherUserID`
  - `Manager`
  - `NewManager`
  - `CreateChat`
  - `EndChat`
  - `SendMessage`
  - `GetChat`
  - `RemoveChat`
  - `Message`
- `flamingo_chat/internal/matchmaking`
  - `FindMatch`
  - `QueueEntry`
  - `Queue`
  - `NewQueue`
  - `AddUser`
  - `RemoveUser`
  - `RemoveUsers`
- `flamingo_chat/internal/user`
  - `Session`
  - `User`
- `flamingo_chat/internal/server`
  - `PresenceStats`
  - `GetPresenceStats`
  - `Server`
  - `NewServer`
  - `CreateUser`
  - `SetPreferences`
  - `SetOnline`
  - `JoinQueue`
  - `TryMatch`
  - `SendMessage`
  - `CreateSession`
  - `PingSession`
  - `DisconnectSession`
  - `LeaveChat`
  - `State`
  - `StartCleanupWorker`
  - `cleanupInactiveSessions`
- `flamingo_chat/internal/transport/websocket`

## Notes

- This README is intended as a quick overview of the repository structure.
- Implementation details are organized under `internal/` to keep the server modular and testable.
- The `cmd/server` package contains the main application entry point.



Browser → WebSocket bytes
  → client.readPump() reads bytes
  → app.HandleMessage() called
  → events.ParseEvent() parses JSON
  → switch on event.Type
  → server.SomeMethod() does the work
  → server callback fires (OnMatchFound etc.)
  → hub.SendToUser() finds the right client
  → client.Send() puts it on the send channel
  → client.writePump() sends it over WebSocket
  → Browser receives bytes