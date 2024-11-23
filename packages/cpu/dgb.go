package cpu

import (
	"pajalic.go.emulator/packages/pubsub"
	"strings"

	log "pajalic.go.emulator/packages/logger"
)

var dbgMsg [1024]byte
var msgSize = 0

func DbgUpdate() {
	if pubsub.BusCtx().BusRead(0xFF02) == 0x81 {
		var c = pubsub.BusCtx().BusRead(0xFF01)

		if msgSize < len(dbgMsg) {
			dbgMsg[msgSize] = c
			msgSize++
		} else {
			log.Warn("dbgMsg buffer overflow")
			msgSize = 0 // Reset to avoid further errors
		}

		pubsub.BusCtx().BusWrite(0xFF02, 0)
	}
}

func DbgPrint() bool {
	if msgSize > 0 {
		debugmsg := string(dbgMsg[:msgSize])
		log.Info("DBG: %s", debugmsg)

		msgSize = 0 // Reset msgSize after printing

		if strings.Contains(debugmsg, "F") {
			return false
		}
	}
	return true
}
