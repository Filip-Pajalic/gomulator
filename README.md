# GoMulator
Game Boy emulator written in Go

## Usage

### Basic usage:
```bash
./gomulator <rom_file>
```

### With options:
```bash
# Skip the boot animation
./gomulator -skip-boot rom.gb

# Enable debug mode
./gomulator -debug rom.gb

# Combine options
./gomulator -skip-boot -debug rom.gb
```

## Controls

### Game Controls:
- **Arrow keys**: D-pad
- **Z**: B button  
- **X**: A button
- **Enter**: Start
- **Tab**: Select

### Boot Animation:
- **Space/Enter/Escape**: Skip boot animation

## Build

```bash
go build -o gomulator.exe ./cmd/main.go
```

## Command Line Options

- `-skip-boot`: Skip the Nintendo logo boot animation
- `-debug`: Enable debug mode with verbose logging