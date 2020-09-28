package service

import (
	"errors"
	"fmt"
	"grpc_youtube_tutorial/pb"
	"sync"

	"github.com/jinzhu/copier"
)

// ErrAlreadyExists for records that already exists
var ErrAlreadyExists = errors.New("record already exists")

// LaptopStore is an interface for laptop storage
type LaptopStore interface {
	Save(laptop *pb.Laptop) error
	Find(laptopID string) (*pb.Laptop, bool)
	Search(filter *pb.Filter, found func(laptop *pb.Laptop) error) error
}

// InMemoryLaptopStore in-memory laptop storage
type InMemoryLaptopStore struct {
	mutex sync.RWMutex
	data  map[string]*pb.Laptop
}

// NewInMemoryLaptopStore returns a InMemoryLaptopStore
func NewInMemoryLaptopStore() *InMemoryLaptopStore {
	return &InMemoryLaptopStore{
		data: make(map[string]*pb.Laptop),
	}
}

// Save laptop to memory store
func (store *InMemoryLaptopStore) Save(laptop *pb.Laptop) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if store.data[laptop.Id] != nil {
		return ErrAlreadyExists
	}

	// deep copy
	other := &pb.Laptop{}
	err := copier.Copy(other, laptop)
	if err != nil {
		return fmt.Errorf("unable to copy laptop: %v", err)
	}

	store.data[other.Id] = other

	return nil
}

// Find checks if laptop with the Id of laptopID is in the store
func (store *InMemoryLaptopStore) Find(laptopID string) (*pb.Laptop, bool) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	laptop := store.data[laptopID]

	if laptop == nil {
		return nil, false
	}

	other, err := deepCopy(laptop)
	if err != nil {
		return nil, false
	}

	return other, true
}

// Search takes a filter and a callback function which will be called if laptop(s) are found
func (store *InMemoryLaptopStore) Search(filter *pb.Filter, found func(laptop *pb.Laptop) error) error {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	for _, laptop := range store.data {
		if isQualified(filter, laptop) {
			other, err := deepCopy(laptop)
			if err != nil {
				return err
			}
			err = found(other)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func isQualified(filter *pb.Filter, laptop *pb.Laptop) bool {
	if laptop.GetPriceUsd() > filter.GetMaxPriceUsd() {
		return false
	}

	if laptop.GetCpu().GetNumberCores() < filter.GetMinCpuCores() {
		return false
	}

	if float64(laptop.GetCpu().GetMinGhz()) < filter.GetMinCpuGhz() {
		return false
	}

	if toBit(laptop.GetRam()) < toBit(filter.GetMinMemory()) {
		return false
	}

	return true

}

func toBit(memory *pb.Memory) uint64 {
	value := uint64(memory.GetValue())

	switch memory.GetUnit() {
	case pb.Memory_BIT:
		return value
	case pb.Memory_BYTE:
		return value << 3 // 8 = 2^3
	case pb.Memory_KILOBYTE:
		return value << 13 // 1024 = 2^10 8 = 2^3
	case pb.Memory_MEGABYTE:
		return value << 23
	case pb.Memory_GIGABYTE:
		return value << 33
	case pb.Memory_TERABYTE:
		return value << 43
	default:
		return 0
	}
}

func deepCopy(laptop *pb.Laptop) (*pb.Laptop, error) {
	other := &pb.Laptop{}
	err := copier.Copy(other, laptop)
	if err != nil {
		return nil, err
	}
	return other, nil
}
