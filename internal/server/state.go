package server

import "flamingo_chat/internal/user"

type State struct {
	Users    map[string]*user.User
	Sessions map[string]*user.Session
}