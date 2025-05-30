package memory

import (
	"app/internal/logger"
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
	Read(address uint16) byte
	Write(address uint16, value byte)
}

type Ppu interface {
	OamRead(address uint16) byte
	OamWrite(address uint16, value byte)
	WramWrite(address uint16, value byte)
	WramRead(address uint16) byte
}

type Bus struct {
	cart       Cart
	ram        Ram
	dma        Dma
	ppu        Ppu
	io         IO
	IERegister byte
	IFRegister byte
}

var busInstance *Bus

func NewBus(cart Cart, ram Ram, dma Dma, ppu Ppu, io IO) *Bus {
	busInstance = &Bus{
		cart: cart,
		ram:  ram,
		dma:  dma,
		ppu:  ppu,
		io:   io,
	}
	return busInstance
}

// BusCtx returns the singleton Bus instance
func BusCtx() *Bus {
	return busInstance
}

// BusRead reads a byte from the bus at the specified address
func (b *Bus) BusRead(address uint16) byte {
	switch {
	case address < 0x8000:
		// Cartridge ROM
		return b.cart.CartRead(address)
	case address < 0xA000:
		// Video RAM (VRAM)
		return b.ppu.WramRead(address)
	case address < 0xC000:
		// External RAM
		return b.cart.CartRead(address)
	case address < 0xE000:
		// Work RAM (WRAM)
		return b.ram.WramRead(address)
	case address < 0xFE00:
		// Echo RAM (not used, mirrors WRAM)
		return 0 // Return 0 for unused memory
	case address < 0xFEA0:
		// Sprite Attribute Table (OAM)
		if b.dma.DMATransferring() {
			// During DMA transfer, OAM cannot be accessed
			return 0xFF
		}
		return b.ppu.OamRead(address)
	case address < 0xFF00:
		// Unusable memory
		return 0xFF
	case address < 0xFF80:
		// I/O Registers
		return b.io.Read(address)
	case address < 0xFFFF:
		// High RAM (HRAM)
		return b.ram.HramRead(address)
	case address == 0xFFFF:
		// Interrupt Enable Register
		return b.IERegister
	default:
		logger.Warn("BusRead: Invalid address %04X", address)
		return 0xFF
	}
}

// BusWrite writes a byte to the bus at the specified address
func (b *Bus) BusWrite(address uint16, data byte) {
	switch {
	case address < 0x8000:
		// Cartridge ROM (writing may affect memory bank controllers)
		b.cart.CartWrite(address, data)
	case address < 0xA000:
		// Video RAM (VRAM)
		b.ppu.WramWrite(address, data)
	case address < 0xC000:
		// External RAM
		b.cart.CartWrite(address, data)
	case address < 0xE000:
		// Work RAM (WRAM)
		b.ram.WramWrite(address, data)
	case address < 0xFE00:
		// Echo RAM (not used, mirrors WRAM)
		// Writes are ignored
	case address < 0xFEA0:
		// Sprite Attribute Table (OAM)
		if b.dma.DMATransferring() {
			// During DMA transfer, OAM cannot be accessed
			return
		}
		b.ppu.OamWrite(address, data)
	case address < 0xFF00:
		// Unusable memory
		// Writes are ignored
	case address < 0xFF80:
		// I/O Registers
		b.io.Write(address, data)
	case address < 0xFFFF:
		// High RAM (HRAM)
		b.ram.HramWrite(address, data)
	case address == 0xFFFF:
		// Interrupt Enable Register
		b.IERegister = data
	default:
		logger.Warn("BusWrite: Invalid address %04X", address)
	}
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

func (b *Bus) GetInterruptEnable() byte {
	return b.IERegister
}

func (b *Bus) SetInterruptEnable(value byte) {
	b.IERegister = value
}

func (b *Bus) GetInterruptFlags() byte {
	return b.IFRegister
}

func (b *Bus) SetInterruptFlags(value byte) {
	b.IFRegister = value
}
