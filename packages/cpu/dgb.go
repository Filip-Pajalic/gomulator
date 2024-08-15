package cpu

import (
	"pajalic.go.emulator/packages/memory"
	"strings"

	log "pajalic.go.emulator/packages/logger"
)

var dbgMsg [1024]byte
var msgSize = 0

func DbgUpdate() {
	if memory.BusRead(0xFF02) == 0x81 {
		var c = memory.BusRead(0xFF01)

		dbgMsg[msgSize] = c
		msgSize++

		memory.BusWrite(0xFF02, 0)
	}
}

func DbgPrint() bool {
	if dbgMsg[0] != 0 {
		log.Info("DBG: %s\n", dbgMsg)

	}
	debugmsg := string(dbgMsg[:])
	if strings.Contains(debugmsg, "F") {
		return false
	}
	return true

}
