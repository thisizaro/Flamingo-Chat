# Flamingo Chat

## File Structure

```
│   go.mod
│   go.sum
│
├───cmd
│   └───server
│       main.go
│
└───internal
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
    │       client.go
    │       handler.go
    │       hub.go
    │       router.go
    │
    ├───user
    │   session.go
    │   user.go
    │
    └───utils
```

## Code Base

- **cmd/server/main.go**
  - main

- **internal/chat/chat.go**
  - GetOtherUserID

- **internal/chat/manager.go**
  - NewManager
  - CreateChat
  - EndChat
  - SendMessage
  - GetChat
  - RemoveChat

- **internal/chat/message.go**

- **internal/events/builder.go**
  - BuildEvent

- **internal/events/events.go**

- **internal/events/parser.go**
  - ParseEvent

- **internal/matchmaking/matcher.go**
  - FindMatch

- **internal/matchmaking/queue.go**
  - NewQueue
  - AddUser
  - RemoveUser
  - RemoveUsers

- **internal/server/presence.go**
  - GetPresenceStats

- **internal/server/server.go**
  - NewServer
  - CreateUser
  - SetPreferences
  - SetOnline
  - JoinQueue
  - TryMatch
  - SendMessage
  - CreateSession
  - PingSession
  - DisconnectSession
  - LeaveChat

- **internal/server/state.go**

- **internal/server/worker.go**
  - StartCleanupWorker
  - cleanupInactiveSessions

- **internal/transport/websocket/client.go**
  - ReadPump
  - WritePump

- **internal/transport/websocket/handler.go**
  - ServeWS

- **internal/transport/websocket/hub.go**
  - NewHub
  - Run
  - GetClientByID
  - SendToClient
  - Broadcast
  - BroadcastPresence

- **internal/transport/websocket/router.go**
  - HandleEvent
  - handleJoinQueue
  - handleSendMessage
  - handleNextPartner

- **internal/user/session.go**

- **internal/user/user.go**


 go run .\allStructus.go
Package: flamingo_chat
  Function: main
  Function: main
Package: flamingo_chat/internal/events
  Function: BuildEvent
  Struct: Event
  Struct: JoinQueuePayload
  Struct: SendMessagePayload
  Struct: MatchFoundPayload
  Struct: MessageReceivedPayload
  Struct: ErrorPayload
  Struct: NextPartnerPayload
  Struct: ChatEndedPayload
  Function: ParseEvent
Package: flamingo_chat/internal/chat
  Struct: Chat
  Method: GetOtherUserID
  Struct: Manager
  Function: NewManager
  Method: CreateChat
  Method: EndChat
  Method: SendMessage
  Method: GetChat
  Method: RemoveChat
  Struct: Message
Package: flamingo_chat/internal/matchmaking
  Method: FindMatch
  Struct: QueueEntry
  Struct: Queue
  Function: NewQueue
  Method: AddUser
  Method: RemoveUser
  Method: RemoveUsers
Package: flamingo_chat/internal/user
  Struct: Session
  Struct: User
Package: flamingo_chat/internal/server
  Struct: PresenceStats
  Method: GetPresenceStats
  Struct: Server
  Function: NewServer
  Method: CreateUser
  Method: SetPreferences
  Method: SetOnline
  Method: JoinQueue
  Method: TryMatch
  Method: SendMessage
  Method: CreateSession
  Method: PingSession
  Method: DisconnectSession
  Method: LeaveChat
  Struct: State
  Method: StartCleanupWorker
  Method: cleanupInactiveSessions
Package: flamingo_chat/internal/transport/websocket
  Struct: Client
  Method: ReadPump
  Method: WritePump
  Function: ServeWS
  Struct: Hub
  Function: NewHub
  Method: Run
  Method: GetClientByID
  Method: SendToClient
  Method: Broadcast
  Method: BroadcastPresence
  Function: HandleEvent
  Function: handleJoinQueue
  Function: handleSendMessage
  Function: handleNextPartner
Package: flamingo_chat/cmd/server
  Function: main