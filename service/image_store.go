package service

import (
	"bytes"
	"fmt"
	"os"
	"sync"

	"github.com/google/uuid"
)

// Maximage size is 1 mb
const maxImageSize = 1 << 20

// ImageStore is an interface to store laptop images
type ImageStore interface {
	Save(laptopID string, imageType string, image bytes.Buffer) (string, error)
}

// DiskImageStore stores image on disk and it's information on memory
type DiskImageStore struct {
	mutex       sync.RWMutex
	imageFolder string
	images      map[string]*ImageInfo
}

// ImageInfo contains information about the laptop image
type ImageInfo struct {
	LaptopID string
	Type     string
	Path     string
}

// NewDiskImageStore returns a new image store
func NewDiskImageStore(imageFolder string) *DiskImageStore {
	return &DiskImageStore{
		imageFolder: imageFolder,
		images:      make(map[string]*ImageInfo),
	}
}

// Save methods saves given file on disk
func (store *DiskImageStore) Save(laptopID string, imageType string, image bytes.Buffer) (string, error) {
	imageID, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("cannot create image id: %v", err)
	}

	imagePath := fmt.Sprintf("%s/%s%s", store.imageFolder, imageID, imageType)

	file, err := os.Create(imagePath)
	if err != nil {
		return "", fmt.Errorf("cannot create image file: %v", err)
	}

	_, err = image.WriteTo(file)
	if err != nil {
		return "", fmt.Errorf("cannot create image err: %v", err)
	}

	store.mutex.Lock()
	defer store.mutex.Unlock()

	store.images[imageID.String()] = &ImageInfo{
		LaptopID: laptopID,
		Type:     imageType,
		Path:     imagePath,
	}

	return imageID.String(), nil
}
