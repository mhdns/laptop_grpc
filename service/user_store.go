package service

import (
	"fmt"
	"sync"
)

// UserStore is a interface to store users
type UserStore interface {
	// Save saves the user to the store
	Save(user *User) error
	// Find finds the user with matching username
	Find(username string) (*User, error)
}

// InMemoryUserStore stores users in memory
type InMemoryUserStore struct {
	mutex sync.RWMutex
	users map[string]*User
}

// NewInMemoryUserStore return a pointer to a InMemoryUserStore
func NewInMemoryUserStore() *InMemoryUserStore {
	return &InMemoryUserStore{
		users: make(map[string]*User),
	}
}

// Save saves the user to the store
func (store *InMemoryUserStore) Save(user *User) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if store.users[user.Username] != nil {
		return fmt.Errorf("user already exists")
	}

	store.users[user.Username] = user.Clone()

	return nil
}

// Find finds the user with matching username
func (store *InMemoryUserStore) Find(username string) (*User, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	user := store.users[username]

	if user == nil {
		return nil, nil
	}

	return user.Clone(), nil
}
