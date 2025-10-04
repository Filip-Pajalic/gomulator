package cpu

import (
	"app/internal/logger"
	"app/internal/memory"
)

type BootRomContext struct {
	BootRomEnabled bool
}

var bootRomInstance *BootRomContext

func NewBootRomContext() *BootRomContext {
	logger.Debug("Boot ROM: Initializing boot ROM context")
	return &BootRomContext{
		BootRomEnabled: true,
	}
}

func BootRomCtx() *BootRomContext {
	if bootRomInstance == nil {
		bootRomInstance = NewBootRomContext()
	}
	return bootRomInstance
}

func (b *BootRomContext) IsBootRomEnabled() bool {
	return b.BootRomEnabled
}

func (b *BootRomContext) DisableBootRom() {
	logger.Debug("Boot ROM: Disabling boot ROM")
	b.BootRomEnabled = false
}

func (b *BootRomContext) SimulateBootSequence() {
	logger.Info("Boot ROM: Simulating boot sequence")

	b.loadNintendoLogoToVRAM()

	cpu := CpuCtx()
	cpu.Regs.A = 0x01 // DMG boot ROM sets A=01
	cpu.Regs.F = 0xB0 // Z=1, N=0, H=1, C=1
	cpu.Regs.B = 0x00
	cpu.Regs.C = 0x13
	cpu.Regs.D = 0x00
	cpu.Regs.E = 0xD8
	cpu.Regs.H = 0x01
	cpu.Regs.L = 0x4D
	cpu.Regs.Sp = 0xFFFE
	cpu.Regs.Pc = 0x0100 // Start execution at ROM entry point

	logger.Debug("Boot ROM: CPU registers initialized")
	logger.Debug("Boot ROM: A=%02X F=%02X BC=%04X DE=%04X HL=%04X SP=%04X PC=%04X",
		cpu.Regs.A, cpu.Regs.F,
		uint16(cpu.Regs.B)<<8|uint16(cpu.Regs.C),
		uint16(cpu.Regs.D)<<8|uint16(cpu.Regs.E),
		uint16(cpu.Regs.H)<<8|uint16(cpu.Regs.L),
		cpu.Regs.Sp, cpu.Regs.Pc)

	b.initializeIORegisters()

	b.initializeSoundRegisters()

	b.initializeLCDRegisters()

	b.initializeTimerRegisters()

	b.initializeInterruptRegisters()

	// Boot ROM is now disabled
	b.DisableBootRom()

	logger.Info("Boot ROM: Boot sequence complete, hardware initialized")
}

// loadNintendoLogoToVRAM loads the Nintendo logo from cartridge header into VRAM
func (b *BootRomContext) loadNintendoLogoToVRAM() {
	logger.Info("Boot ROM: Loading Nintendo logo from cartridge header to VRAM")

	bus := memory.BusCtx()

	// Nintendo logo starts at 0x0104 in cartridge and is 48 bytes (0x30)
	logoStartAddr := uint16(0x0104)
	logoLength := uint16(0x30) // 48 bytes

	// The boot ROM loads it starting at tile 1 (0x8010), not tile 0
	vramStartAddr := uint16(0x8010)

	logger.Debug("Boot ROM: Reading Nintendo logo from ROM 0x%04X-0x%04X", logoStartAddr, logoStartAddr+logoLength-1)

	for i := uint16(0); i < logoLength; i++ {
		logoByte := bus.BusRead(logoStartAddr + i)
		vramAddr := vramStartAddr + i

		bus.BusWrite(vramAddr, logoByte)

		if i < 8 { // Log first few bytes for debugging
			logger.Debug("Boot ROM: Logo byte[%d] = 0x%02X -> VRAM[0x%04X]", i, logoByte, vramAddr)
		}
	}

	// Also set up basic background tile map to show the logo
	// Boot ROM typically sets up a simple tile map that references these tiles
	bgMapAddr := uint16(0x9800) // Background tile map start

	// This creates a basic test pattern using the loaded logo tiles
	for row := 0; row < 4; row++ {
		for col := 0; col < 12; col++ {
			tileIndex := uint8(1 + (row*12 + col)) // Start from tile 1
			mapAddr := bgMapAddr + uint16(row*32+col)
			if tileIndex <= 3 { // Only use first few tiles to avoid issues
				bus.BusWrite(mapAddr, tileIndex)
				if row == 0 && col < 4 {
					logger.Debug("Boot ROM: Map[%d,%d] = tile %d at 0x%04X", row, col, tileIndex, mapAddr)
				}
			}
		}
	}

	logger.Info("Boot ROM: Nintendo logo loaded to VRAM successfully")
}

func (b *BootRomContext) initializeIORegisters() {
	logger.Debug("Boot ROM: Initializing I/O registers")

	// P1/JOYP (FF00) - Joypad register
	memory.BusCtx().BusWrite(0xFF00, 0xCF)

	// Serial data (FF01-FF02)
	memory.BusCtx().BusWrite(0xFF01, 0x00) // Serial transfer data
	memory.BusCtx().BusWrite(0xFF02, 0x7E) // Serial transfer control

	// Divider register (FF04) - initialized by timer
	memory.BusCtx().BusWrite(0xFF04, 0x18) // DIV register (continuously incrementing)

	// Sound registers will be handled separately

	// Boot ROM disable (FF50) - will be set when we disable boot ROM
	memory.BusCtx().BusWrite(0xFF50, 0x01) // Boot ROM disabled

	logger.Debug("Boot ROM: I/O registers initialized")
}

func (b *BootRomContext) initializeSoundRegisters() {
	logger.Debug("Boot ROM: Initializing sound registers")

	// Sound Channel 1 (FF10-FF14)
	memory.BusCtx().BusWrite(0xFF10, 0x80) // NR10
	memory.BusCtx().BusWrite(0xFF11, 0xBF) // NR11
	memory.BusCtx().BusWrite(0xFF12, 0xF3) // NR12
	memory.BusCtx().BusWrite(0xFF14, 0xBF) // NR14

	// Sound Channel 2 (FF16-FF19)
	memory.BusCtx().BusWrite(0xFF16, 0x3F) // NR21
	memory.BusCtx().BusWrite(0xFF17, 0x00) // NR22
	memory.BusCtx().BusWrite(0xFF19, 0xBF) // NR24

	// Sound Channel 3 (FF1A-FF1E)
	memory.BusCtx().BusWrite(0xFF1A, 0x7F) // NR30
	memory.BusCtx().BusWrite(0xFF1B, 0xFF) // NR31
	memory.BusCtx().BusWrite(0xFF1C, 0x9F) // NR32
	memory.BusCtx().BusWrite(0xFF1E, 0xBF) // NR34

	// Sound Channel 4 (FF20-FF23)
	memory.BusCtx().BusWrite(0xFF20, 0xFF) // NR41
	memory.BusCtx().BusWrite(0xFF21, 0x00) // NR42
	memory.BusCtx().BusWrite(0xFF22, 0x00) // NR43
	memory.BusCtx().BusWrite(0xFF23, 0xBF) // NR44

	// Sound Control (FF24-FF26)
	memory.BusCtx().BusWrite(0xFF24, 0x77) // NR50
	memory.BusCtx().BusWrite(0xFF25, 0xF3) // NR51
	memory.BusCtx().BusWrite(0xFF26, 0xF1) // NR52

	// Wave Pattern RAM (FF30-FF3F) - initialized to specific pattern
	wavePattern := []byte{
		0x84, 0x40, 0x43, 0xAA, 0x2D, 0x78, 0x92, 0x3C,
		0x60, 0x59, 0x59, 0xB0, 0x34, 0xB8, 0x2E, 0xDA,
	}
	for i, val := range wavePattern {
		memory.BusCtx().BusWrite(0xFF30+uint16(i), val)
	}

	logger.Debug("Boot ROM: Sound registers initialized")
}

func (b *BootRomContext) initializeLCDRegisters() {
	logger.Debug("Boot ROM: Initializing LCD/PPU registers")

	// LCD Control (FF40) - LCDC
	memory.BusCtx().BusWrite(0xFF40, 0x91) // LCD enabled, BG on, sprites on, window off

	// LCD Status (FF41) - STAT
	memory.BusCtx().BusWrite(0xFF41, 0x85) // Mode 1 (V-Blank), coincidence flag

	// Scroll registers (FF42-FF43)
	memory.BusCtx().BusWrite(0xFF42, 0x00) // SCY - scroll Y
	memory.BusCtx().BusWrite(0xFF43, 0x00) // SCX - scroll X

	// LY (FF44) - LCD Y coordinate
	memory.BusCtx().BusWrite(0xFF44, 0x91) // Current scanline (in V-blank)

	// LYC (FF45) - LY compare
	memory.BusCtx().BusWrite(0xFF45, 0x00) // LY compare value

	// DMA (FF46) - DMA transfer
	memory.BusCtx().BusWrite(0xFF46, 0xFF) // No DMA transfer active

	// Palette registers (FF47-FF49)
	memory.BusCtx().BusWrite(0xFF47, 0xFC) // BGP - background palette
	memory.BusCtx().BusWrite(0xFF48, 0xFF) // OBP0 - object palette 0
	memory.BusCtx().BusWrite(0xFF49, 0xFF) // OBP1 - object palette 1

	// Window position (FF4A-FF4B)
	memory.BusCtx().BusWrite(0xFF4A, 0x00) // WY - window Y position
	memory.BusCtx().BusWrite(0xFF4B, 0x00) // WX - window X position

	logger.Debug("Boot ROM: LCD/PPU registers initialized")
}

func (b *BootRomContext) initializeTimerRegisters() {
	logger.Debug("Boot ROM: Initializing timer registers")

	// Timer registers (FF05-FF07)
	memory.BusCtx().BusWrite(0xFF05, 0x00) // TIMA - timer counter
	memory.BusCtx().BusWrite(0xFF06, 0x00) // TMA - timer modulo
	memory.BusCtx().BusWrite(0xFF07, 0xF8) // TAC - timer control

	logger.Debug("Boot ROM: Timer registers initialized")
}

func (b *BootRomContext) initializeInterruptRegisters() {
	logger.Debug("Boot ROM: Initializing interrupt registers")

	// Interrupt registers
	memory.BusCtx().BusWrite(0xFF0F, 0xE1) // IF - interrupt flag
	memory.BusCtx().BusWrite(0xFFFF, 0x00) // IE - interrupt enable

	logger.Debug("Boot ROM: Interrupt registers initialized")
}

func (b *BootRomContext) ReadBootRom(address uint16) byte {
	if !b.BootRomEnabled {
		// Boot ROM disabled, let cartridge handle the read
		return 0xFF // Will be handled by cartridge
	}

	// Since we simulate the boot sequence, we don't need actual boot ROM data
	// Just return 0x00 for any boot ROM reads that might occur
	logger.Debug("Boot ROM: Read from boot ROM area %04X (simulated)", address)
	return 0x00
}
