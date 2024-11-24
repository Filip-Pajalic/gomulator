package memory

import (
	"sync"
)

type Ram interface {
	WramRead(address uint16) byte
	WramWrite(address uint16, value byte)
	HramRead(address uint16) byte
	HramWrite(address uint16, value byte)
}

type Cart interface {
	CartRead(address uint16) byte
	CartWrite(address uint16, data byte)
}

type Dma interface {
	DMATransferring() bool
}

type Cpu interface {
	GetIERegister() byte
	SetIERegister(n byte)
}

type IO interface {
	IoRead(address uint16) byte
}

type Ppu interface {
	OamRead(address uint16) byte
	OamWrite(address uint16, value byte)
	PpuWramWrite(address uint16, value byte)
	PpuWramRead(address uint16) byte
}

type Bus struct {
	cart       Cart
	ram        Ram
	dma        Dma
	ppu        Ppu
	IERegister byte
}

var busInstance *Bus
var once sync.Once

func NewBus(cart Cart, ram Ram, dma Dma, ppu Ppu) *Bus {
	busInstance = &Bus{
		cart: cart,
		ram:  ram,
		dma:  dma,
		ppu:  ppu,
	}
	return busInstance
}

// BusCtx returns the singleton Bus instance
func BusCtx() *Bus {
	return busInstance
}

// BusRead reads a byte from the bus at the specified address
func (b *Bus) BusRead(address uint16) byte {
	if address < 0x8000 {
		return b.cart.CartRead(address)
	} else if address < 0xA000 {
		return b.ppu.PpuWramRead(address)
	} else if address < 0xC000 {
		return b.cart.CartRead(address)
	} else if address < 0xE000 {
		return b.ram.WramRead(address)
	} else if address < 0xFE00 {
		// Reserved echo RAM, not used
		return 0
	} else if address < 0xFEA0 {
		if b.dma.DMATransferring() {
			return 0xFF
		}
		return b.ppu.OamRead(address)
	} else if address < 0xFF00 {
		return 0
	} else if address < 0xFF80 {
		//return cpu.IoRead(address)
	} else if address == 0xFFFF {
		return b.IERegister
	}
	return b.ram.HramRead(address)
}

// BusWrite writes a byte to the bus at the specified address
func (b *Bus) BusWrite(address uint16, data byte) {
	if address < 0x8000 {
		b.cart.CartWrite(address, data)
	} else if address < 0xA000 {
		b.ppu.PpuWramWrite(address, data)
	} else if address < 0xC000 {
		b.cart.CartWrite(address, data)
	} else if address < 0xE000 {
		b.ram.WramWrite(address, data)
	} else if address < 0xFE00 {
		// Reserved echo RAM, not used
		return
	} else if address < 0xFEA0 {
		if b.dma.DMATransferring() {
			return
		}
		b.ppu.OamWrite(address, data)
	} else if address < 0xFF00 {
		// Reserved, not used
		return
	} else if address < 0xFF80 {
		//cpu.IoWrite(address, data)
	} else if address == 0xFFFF {
		// Interrupt Enable Register
		b.IERegister = data
	}
	b.ram.HramWrite(address, data)
}

// BusRead16 reads two bytes from the bus starting at the specified address
func (b *Bus) BusRead16(address uint16) uint16 {
	lo := uint16(b.BusRead(address))
	hi := uint16(b.BusRead(address + 1))
	return lo | (hi << 8)
}

// BusWrite16 writes two bytes to the bus starting at the specified address
func (b *Bus) BusWrite16(address uint16, data uint16) {
	b.BusWrite(address, byte(data&0xFF))
	b.BusWrite(address+1, byte((data>>8)&0xFF))
}
