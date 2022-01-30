package main

import (
	"os"

	gameboypackage "pajalic.go.emulator/packages"
)

func main() {
	gameboypackage.Emu_run(len(os.Args), os.Args)
}
