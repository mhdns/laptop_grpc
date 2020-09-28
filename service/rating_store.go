package service

import "sync"

// RatingStore is an interface to store laptop ratings
type RatingStore interface {
	Add(laptopID string, score float64) (*Rating, error)
}

// Rating contains the rating information for given laptop
type Rating struct {
	Count uint32
	Sum   float64
}

// InMemoryRatingStore stores laptop ratings inMemory
type InMemoryRatingStore struct {
	mutex  sync.RWMutex
	rating map[string]*Rating
}

// NewInMemoryRatingStore returns a InMemoryRatingStore
func NewInMemoryRatingStore() *InMemoryRatingStore {
	return &InMemoryRatingStore{
		rating: make(map[string]*Rating),
	}
}

// Add adds a rating to the Rating store
func (store *InMemoryRatingStore) Add(laptopID string, score float64) (*Rating, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	rating := store.rating[laptopID]
	if rating == nil {
		rating = &Rating{
			Count: 1,
			Sum:   score,
		}
	} else {
		rating.Count++
		rating.Sum += score
	}

	store.rating[laptopID] = rating
	return store.rating[laptopID], nil
}
