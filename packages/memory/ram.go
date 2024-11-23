package memory

import (
	"sync"

	log "pajalic.go.emulator/packages/logger"
	"pajalic.go.emulator/packages/pubsub"
)

// RamContext represents the state of WRAM and HRAM
type RamContext struct {
	Wram [0x2000]byte // 8KB WRAM (0xC000 - 0xDFFF)
	Hram [0x80]byte   // 128B HRAM (0xFF80 - 0xFFFE)

	mu sync.RWMutex // Mutex to protect memory access
}

// singleton instance of RamContext
var ramInstance *RamContext

// GetRamContext returns the singleton RamContext instance
func GetRamContext() *RamContext {
	if ramInstance == nil {
		ramInstance = &RamContext{}
	}
	return ramInstance
}

// WramRead reads a byte from WRAM at the given address
func (r *RamContext) WramRead(address uint16) byte {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if address < 0xC000 || address >= 0xE000 {
		log.Warn("WRAM Read: Invalid address %04X", address)
		return 0xFF // Return default value for invalid addresses
	}
	return r.Wram[address-0xC000]
}

// WramWrite writes a byte to WRAM at the given address
func (r *RamContext) WramWrite(address uint16, value byte) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if address < 0xC000 || address >= 0xE000 {
		log.Warn("WRAM Write: Invalid address %04X", address)
		return
	}
	r.Wram[address-0xC000] = value
}

// HramRead reads a byte from HRAM at the given address
func (r *RamContext) HramRead(address uint16) byte {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if address < 0xFF80 || address >= 0xFFFF {
		log.Warn("HRAM Read: Invalid address %04X", address)
		return 0xFF // Return default value for invalid addresses
	}
	return r.Hram[address-0xFF80]
}

// HramWrite writes a byte to HRAM at the given address
func (r *RamContext) HramWrite(address uint16, value byte) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if address < 0xFF80 || address >= 0xFFFF {
		log.Warn("HRAM Write: Invalid address %04X", address)
		return
	}
	r.Hram[address-0xFF80] = value
}

// InitializeMemory sets up ReadWriteConfig for WRAM and HRAM and starts processing
func InitializeMemory() {
	ram := GetRamContext()

	// Configure WRAM Read and Write events
	wramConfig := pubsub.NewReadWriteConfig[uint16, byte](
		pubsub.MemoryWramReadEvent,
		pubsub.MemoryWramWriteEvent,
		ram.WramRead,
		ram.WramWrite,
	)

	// Start processing WRAM transactions
	go pubsub.ProcessChannelTransactions(wramConfig)

	// Configure HRAM Read and Write events
	hramConfig := pubsub.NewReadWriteConfig[uint16, byte](
		pubsub.MemoryHramReadEvent,
		pubsub.MemoryHramWriteEvent,
		ram.HramRead,
		ram.HramWrite,
	)

	// Start processing HRAM transactions
	go pubsub.ProcessChannelTransactions(hramConfig)
}
