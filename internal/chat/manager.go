package chat

import (
	"fmt"
	"time"
)

type Manager struct {
	chats map[string]*Chat
}

func NewManager() *Manager {
	return &Manager{
		chats: make(map[string]*Chat),
	}
}

// create a new chat
func (m *Manager) CreateChat(chatID string, userIDs []string) *Chat {
	chat := &Chat{
		ID:        chatID,
		UserIDs:   userIDs,
		CreatedAt: time.Now(),
		Status:    StatusActive,
	}

	m.chats[chatID] = chat
	return chat
}

// end a chat
func (m *Manager) EndChat(chatID string) {
	if chat, ok := m.chats[chatID]; ok {
		chat.Status = StatusEnded
	}
}

// send message:
func (m *Manager) SendMessage(
	chatID,
	senderID,
	content string,
	messageType MessageType,
) *Message {

	chat, ok := m.chats[chatID]
	if !ok {
		fmt.Println("[CHAT] Chat not found:", chatID)
		return nil
	}

	if chat.Status != StatusActive {
		fmt.Println("[CHAT] Chat inactive:", chatID)
		return nil
	}

	msg := &Message{
		ID:          fmt.Sprintf("%d", time.Now().UnixNano()),
		ChatID:      chatID,
		SenderID:    senderID,
		Content:     content,
		MessageType: messageType,
		CreatedAt:   time.Now(),
	}

	chat.Messages = append(chat.Messages, msg)

	fmt.Println("[CHAT] Message sent")
	fmt.Println("  ChatID:", chatID)
	fmt.Println("  Sender:", senderID)
	fmt.Println("  Type:", messageType)
	fmt.Println("  Content:", content)

	return msg
}

func (m *Manager) GetChat(
	chatID string,
) (*Chat, bool) {

	chat, ok := m.chats[chatID]

	return chat, ok
}

func (m *Manager) RemoveChat(
	chatID string,
) {

	delete(m.chats, chatID)

	fmt.Println("[CHAT] Chat removed")
	fmt.Println("  ChatID:", chatID)
}
