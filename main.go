package main

import (
	"os"

	"pajalic.go.emulator/packages/emulatorloop"
)

func main() {
	emulatorloop.Run(len(os.Args), os.Args)
}
