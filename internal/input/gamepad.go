package input

type State struct {
	Start  bool
	Select bool
	A      bool
	B      bool
	Up     bool
	Down   bool
	Right  bool
	Left   bool
}

type Context struct {
	ButtonSel  bool
	DirSel     bool
	Controller State
}

var Ctx Context

func Init() {
	Ctx = Context{
		ButtonSel:  false,
		DirSel:     false,
		Controller: State{},
	}
}

func ButtonSel() bool {
	return Ctx.ButtonSel
}

func DirSel() bool {
	return Ctx.DirSel
}

func SetSel(value uint8) {
	// Joypad register uses active-low selection bits: when bit is 0 the group
	// is selected. SetSel receives the written byte and stores booleans that
	// are true when the corresponding group is selected.
	Ctx.ButtonSel = (value & 0x20) == 0
	Ctx.DirSel = (value & 0x10) == 0
}
func GetState() *State {
	return &Ctx.Controller
}

func GetOutput() uint8 {
	var output uint8 = 0xCF

	// When a group is selected (ButtonSel/DirSel true), clear the
	// corresponding bits for pressed buttons (active-low logic on the port).
	if ButtonSel() {
		if GetState().Start {
			output &= ^(uint8(1) << 3)
		}
		if GetState().Select {
			output &= ^(uint8(1) << 2)
		}
		if GetState().A {
			output &= ^(uint8(1) << 0)
		}
		if GetState().B {
			output &= ^(uint8(1) << 1)
		}
	}

	if DirSel() {
		if GetState().Left {
			output &= ^(uint8(1) << 1)
		}
		if GetState().Right {
			output &= ^(uint8(1) << 0)
		}
		if GetState().Up {
			output &= ^(uint8(1) << 2)
		}
		if GetState().Down {
			output &= ^(uint8(1) << 3)
		}
	}

	return output
}
