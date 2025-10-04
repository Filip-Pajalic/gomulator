package cpu

import (
	"app/internal/memory"
	"strings"

	logger "app/internal/logger"
)

var dbgMsg [1024]byte
var msgSize = 0

func DbgUpdate() {
	if memory.BusCtx().BusRead(0xFF02) == 0x81 {
		var c = memory.BusCtx().BusRead(0xFF01)

		if msgSize < len(dbgMsg) {
			dbgMsg[msgSize] = c
			msgSize++
		} else {
			logger.Warn("dbgMsg buffer overflow")
			msgSize = 0 // Reset to avoid further errors
		}

		memory.BusCtx().BusWrite(0xFF02, 0)
	}
}

func DbgPrint() bool {
	if msgSize > 0 {
		debugmsg := string(dbgMsg[:msgSize])
		logger.Info("DBG: %s", debugmsg)

		msgSize = 0 // Reset msgSize after printing

		if strings.Contains(debugmsg, "F") {
			return false
		}
	}
	return true
}
