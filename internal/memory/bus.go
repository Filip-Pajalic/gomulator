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
	VramWrite(address uint16, value byte)
	VramRead(address uint16) byte
}

type Bus struct {
	cart       Cart
	ram        Ram
	dma        Dma
	ppu        Ppu
	io         IO
	cpu        Cpu
	IERegister byte
	IFRegister byte
}

var busInstance *Bus

func NewBus(cart Cart, ram Ram, dma Dma, ppu Ppu, io IO, cpu Cpu) *Bus {
	busInstance = &Bus{
		cart: cart,
		ram:  ram,
		dma:  dma,
		ppu:  ppu,
		io:   io,
		cpu:  cpu,
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
		// Cartridge ROM - but check for boot ROM first
		if address < 0x0100 {
			// Boot ROM area (0x0000-0x00FF)
			// Since we simulate the boot sequence, we always return cartridge data
			// The boot ROM simulation handles initialization, so we skip actual boot ROM reads
		}
		return b.cart.CartRead(address)
	case address < 0xA000:
		// Video RAM (VRAM)
		return b.ppu.VramRead(address)
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
		b.ppu.VramWrite(address, data)
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
		// Interrupt Enable Register - update both bus and CPU
		b.IERegister = data
		if b.cpu != nil {
			b.cpu.SetIERegister(data)
		}
	default:
		logger.Warn("BusWrite: Invalid address %04X", address)
	}
}

// DmaWriteToOam writes directly to OAM, bypassing CPU access restrictions during DMA
func (b *Bus) DmaWriteToOam(address uint16, data byte) {
	if address < 0xFE00 || address >= 0xFEA0 {
		logger.Warn("DMA OAM write: Invalid address %04X", address)
		return
	}
	b.ppu.OamWrite(address, data)
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
