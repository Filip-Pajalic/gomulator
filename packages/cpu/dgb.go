package cpu

import (
	"strings"

	log "pajalic.go.emulator/packages/logger"
)

var dbgMsg [1024]byte
var msgSize = 0

func DbgUpdate() {
	if BusRead(0xFF02) == 0x81 {
		var c = BusRead(0xFF01)

		dbgMsg[msgSize] = c
		msgSize++

		BusWrite(0xFF02, 0)
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
