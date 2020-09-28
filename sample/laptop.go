package sample

import (
	"grpc_youtube_tutorial/pb"

	"github.com/golang/protobuf/ptypes"
)

// NewKeyboard returns a pointer to a keyboard object
func NewKeyboard() *pb.Keyboard {
	return &pb.Keyboard{
		Layout:  randomKeyboardLayout(),
		Backlit: randomBoolean(),
	}
}

// NewCPU returns a pointer to a CPU object
func NewCPU() *pb.CPU {
	brand := randomCPUBrand()
	name := randomCPUNAme(brand)
	numCores := randomInt(2, 8)
	numThreads := randomInt(numCores, 12)
	minGhz := randomFloat32(2.0, 3.5)
	maxGhz := randomFloat32(minGhz, 5.0)

	return &pb.CPU{
		Brand:         brand,
		Name:          name,
		NumberCores:   uint32(numCores),
		NumberThreads: uint32(numThreads),
		MinGhz:        minGhz,
		MaxGhz:        maxGhz,
	}
}

// NewGPU returns a random GPU object
func NewGPU() *pb.GPU {
	gpuBrand := randomGPUBrand()
	gpuName := randomGPUName(gpuBrand)
	minGhz := randomFloat32(1.0, 1.5)
	maxGhz := randomFloat32(minGhz, 2.0)
	memory := randomInt(2, 6)
	return &pb.GPU{
		Brand: gpuBrand,
		Name:  gpuName,
		Memory: &pb.Memory{
			Value: uint32(memory),
			Unit:  pb.Memory_GIGABYTE,
		},
		MinGhz: minGhz,
		MaxGhz: maxGhz,
	}
}

// NewRAM returns a new RAM object
func NewRAM() *pb.Memory {
	memory := randomInt(4, 64)
	return &pb.Memory{
		Value: uint32(memory),
		Unit:  pb.Memory_GIGABYTE,
	}
}

// NewSSD returns a SSD object
func NewSSD() *pb.Storage {
	memory := &pb.Memory{
		Value: uint32(randomInt(128, 1024)),
		Unit:  pb.Memory_GIGABYTE,
	}
	return &pb.Storage{
		Driver: pb.Storage_SSD,
		Memory: memory,
	}
}

// NewHDD returns a SSD object
func NewHDD() *pb.Storage {
	memory := &pb.Memory{
		Value: uint32(randomInt(1, 6)),
		Unit:  pb.Memory_TERABYTE,
	}
	return &pb.Storage{
		Driver: pb.Storage_HDD,
		Memory: memory,
	}
}

// NewScreen returns a screen object
func NewScreen() *pb.Screen {
	return &pb.Screen{
		SizeInch:   randomFloat32(13, 17),
		Resolution: randomScreenResolution(),
		Panel:      randomScreenPanel(),
		Multitouch: randomBoolean(),
	}
}

// NewLaptop returns a screen object
func NewLaptop() *pb.Laptop {
	brand := randomLaptopBrand()
	name := randomLaptopName(brand)
	return &pb.Laptop{
		Id:       randomID(),
		Brand:    brand,
		Name:     name,
		Cpu:      NewCPU(),
		Ram:      NewRAM(),
		Gpus:     []*pb.GPU{NewGPU(), NewGPU()},
		Storages: []*pb.Storage{NewHDD(), NewSSD()},
		Screen:   NewScreen(),
		Keyboard: NewKeyboard(),
		Weight: &pb.Laptop_WeightKg{
			WeightKg: randomFloat64(1.0, 3.0),
		},
		PriceUsd:    randomFloat64(1500, 3000),
		ReleaseYear: uint32(randomInt(2015, 2020)),
		UpdatedAt:   ptypes.TimestampNow(),
	}
}

// RandomLaptopScore generates random laptop rating score
func RandomLaptopScore() float64 {
	return float64(randomInt(1, 10))
}
