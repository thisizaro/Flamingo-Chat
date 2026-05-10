# Flamingo Chat

A simple Go-based chat server with matchmaking, real-time WebSocket transport, and clean modular design.

---

## What this project includes

* Matchmaking system for pairing users in private chats
* WebSocket transport layer for real-time messaging
* Event-based protocol for chat actions
* User session handling and presence tracking
* Background cleanup of inactive sessions

---

## Project Structure

```
│   allFunctions.go
│   allStructus.go
│   FRONTEND.md
│   go.mod
│   go.sum
│   NOTES.md
│   README.md
│   
├───cmd
│   └───server
│           main.go
│           
└───internal
    ├───app
    │       handler.go
    │       
    ├───chat
    │       chat.go
    │       manager.go
    │       message.go
    │       
    ├───events
    │       builder.go
    │       events.go
    │       parser.go
    │       
    ├───integration
    │       integration_test.go
    │       
    ├───matchmaking
    │       matcher.go
    │       matcher_test.go
    │       queue.go
    │       
    ├───server
    │       presence.go
    │       server.go
    │       server_test.go
    │       state.go
    │       worker.go
    │       
    ├───transport
    │   └───websocket
    │           client.go
    │           hub.go
    │           
    ├───user
    │       session.go
    │       user.go
    │       
    └───utils
```

---

## File → Functions Mapping

File: allFunctions.go

* main

File: allStructus.go

* main

File: cmd/server/main.go

* main

File: internal/app/handler.go

* New
* Wire
* HandleMessage
* handleInit
* handleJoinQueue
* handleSendMessage
* handlePing
* handleLeaveChat
* sendError
* unmarshalPayload

File: internal/chat/chat.go

* GetOtherUserID

File: internal/chat/manager.go

* NewManager
* CreateChat
* EndChat
* SendMessage
* GetChat
* RemoveChat

File: internal/chat/message.go

* (message definitions)

File: internal/events/builder.go

* BuildEvent

File: internal/events/events.go

* (event structs)

File: internal/events/parser.go

* ParseEvent

File: internal/integration/integration_test.go

* startTestServer
* connect
* send
* receive
* assertEventType
* TestIntegration_InitNewUser
* TestIntegration_InitReturningUser
* TestIntegration_InvalidUserID_GetsNewID
* TestIntegration_FullMatchAndMessage
* TestIntegration_LeaveChat_PartnerNotified

File: internal/matchmaking/matcher.go

* FindMatch

File: internal/matchmaking/matcher_test.go

* TestFindMatch_BasicPair
* TestFindMatch_NoMatchWrongPreference
* TestFindMatch_EmptyQueue
* TestFindMatch_OneUser
* TestFindMatch_PicksFirstCompatiblePair
* TestRemoveUser
* TestRemoveUsers_Multiple
* contains

File: internal/matchmaking/queue.go

* NewQueue
* AddUser
* RemoveUser
* RemoveUsers

File: internal/server/presence.go

* GetPresenceStats

File: internal/server/server.go

* NewServer
* generateUserID
* InitUser
* SetPreferences
* SetOnline
* JoinQueue
* TryMatch
* SendMessage
* CreateSession
* PingSession
* DisconnectSession
* LeaveChat
* UserExists

File: internal/server/server_test.go

* TestInitUser_NewUser
* TestInitUser_ReturningUser
* TestInitUser_UnknownID_CreatesNew
* TestInitUser_UniqueIDs
* TestJoinQueue_StatusChanges
* TestTryMatch_TwoCompatibleUsers
* TestTryMatch_IncompatibleUsers_NoMatch
* TestLeaveChat_NotifiesPartner
* TestLeaveChat_StatusResetToIdle
* TestSendMessage_DeliveredToPartner

File: internal/server/state.go

* (state definitions)

File: internal/server/worker.go

* StartCleanupWorker
* cleanupInactiveSessions

File: internal/transport/websocket/client.go

* ServeWS
* readPump
* writePump
* Send
* SetUserID
* GetUserID

File: internal/transport/websocket/hub.go

* NewHub
* Run
* RegisterUser
* SendToUser
* SetOnDisconnect

File: internal/user/session.go

* (session struct)

File: internal/user/user.go

* (user struct)

---

## Package Summary

Package: flamingo_chat

* Function: main
* Function: main

Package: flamingo_chat/internal/chat

* Struct: Chat
* Method: GetOtherUserID
* Struct: Manager
* Function: NewManager
* Method: CreateChat
* Method: EndChat
* Method: SendMessage
* Method: GetChat
* Method: RemoveChat
* Struct: Message

Package: flamingo_chat/internal/events

* Function: BuildEvent
* Struct: Event
* Struct: InitPayload
* Struct: JoinQueuePayload
* Struct: SendMessagePayload
* Struct: NextPartnerPayload
* Struct: ReadyPayload
* Struct: MatchFoundPayload
* Struct: MessageReceivedPayload
* Struct: ErrorPayload
* Struct: ChatEndedPayload
* Function: ParseEvent

Package: flamingo_chat/internal/matchmaking

* Method: FindMatch
* Struct: QueueEntry
* Struct: Queue
* Function: NewQueue
* Method: AddUser
* Method: RemoveUser
* Method: RemoveUsers

Package: flamingo_chat/internal/user

* Struct: Session
* Struct: User

Package: flamingo_chat/internal/server

* Struct: PresenceStats
* Method: GetPresenceStats
* Struct: Server
* Function: NewServer
* Function: generateUserID
* Method: InitUser
* Method: SetPreferences
* Method: SetOnline
* Method: JoinQueue
* Method: TryMatch
* Method: SendMessage
* Method: CreateSession
* Method: PingSession
* Method: DisconnectSession
* Method: LeaveChat
* Method: UserExists
* Struct: State
* Method: StartCleanupWorker
* Method: cleanupInactiveSessions

Package: flamingo_chat/internal/transport/websocket

* Struct: Client
* Function: ServeWS
* Method: readPump
* Method: writePump
* Method: Send
* Method: SetUserID
* Method: GetUserID
* Struct: Hub
* Struct: targetedMessage
* Function: NewHub
* Method: Run
* Method: RegisterUser
* Method: SendToUser
* Method: SetOnDisconnect

Package: flamingo_chat/internal/app

* Struct: App
* Function: New
* Method: Wire
* Method: HandleMessage
* Method: handleInit
* Method: handleJoinQueue
* Method: handleSendMessage
* Method: handlePing
* Method: handleLeaveChat
* Method: sendError
* Function: unmarshalPayload

Package: flamingo_chat/cmd/server

* Function: main

Package: flamingo_chat/internal/integration

* (integration tests)

---

## Request Flow

```
Browser → WebSocket bytes
  → client.readPump()
  → app.HandleMessage()
  → events.ParseEvent()
  → switch(event.Type)
  → server.SomeMethod()
  → hub.SendToUser()
  → client.Send()
  → client.writePump()
  → Browser receives bytes
```

---

## Notes

* Organized using Go's `internal/` pattern for modularity
* Clear separation of concerns: transport, server logic, matchmaking, events
* Designed for testability with integration and unit tests
* WebSocket-based real-time architecture
