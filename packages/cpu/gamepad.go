package cpu

type GamePadState struct {
	Start  bool
	Select bool
	A      bool
	B      bool
	Up     bool
	Down   bool
	Right  bool
	Left   bool
}

type GamePadContext struct {
	ButtonSel  bool
	DirSel     bool
	Controller GamePadState
}

var GamePadCtx GamePadContext

func GamePadInit() {
	GamePadCtx = GamePadContext{
		ButtonSel: false,
		DirSel:    false,
		Controller: GamePadState{
			A: false,
			B: false,

			Start:  false,
			Select: false,
			Up:     false,
			Down:   false,
			Left:   false,
			Right:  false,
		},
	}
}

func GamePadButtonSel() bool {
	return GamePadCtx.ButtonSel
}

func GamePadDirSel() bool {
	return GamePadCtx.DirSel
}

func GamepadSetSel(value uint8) {
	GamePadCtx.ButtonSel = (value & 0x20) != 0
	GamePadCtx.DirSel = (value & 0x10) != 0
}
func GamePadGetState() *GamePadState {
	return &GamePadCtx.Controller
}

func GamepadGetOutput() uint8 {
	var output uint8 = 0xCF

	if !GamePadButtonSel() {
		if GamePadGetState().Start {
			output &= ^(uint8(1) << 3)
		}
		if GamePadGetState().Select {
			output &= ^(uint8(1) << 2)
		}
		if GamePadGetState().A {
			output &= ^(uint8(1) << 0)
		}
		if GamePadGetState().B {
			output &= ^(uint8(1) << 1)
		}
	}

	if !GamePadDirSel() {
		if GamePadGetState().Left {
			output &= ^(uint8(1) << 1)
		}
		if GamePadGetState().Right {
			output &= ^(uint8(1) << 0)
		}
		if GamePadGetState().Up {
			output &= ^(uint8(1) << 2)
		}
		if GamePadGetState().Down {
			output &= ^(uint8(1) << 3)
		}
	}

	return output
}
