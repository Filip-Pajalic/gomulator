package main

import (
	"os"
	"pajalic.go.emulator/packages/emulator"
)

func main() {
	emulator.Run(len(os.Args), os.Args)
}
