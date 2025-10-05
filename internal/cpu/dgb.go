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

			// Log character received (for debugging)
			if c >= 32 && c <= 126 { // printable ASCII
				logger.Debug("Serial received: '%c' (0x%02X)", c, c)
			} else {
				logger.Debug("Serial received: 0x%02X", c)
			}
		} else {
			logger.Warn("dbgMsg buffer overflow")
			msgSize = 0 // Reset to avoid further errors
		}

		memory.BusCtx().BusWrite(0xFF02, 0)
	}
}

func DbgPrint() bool {
	if msgSize > 0 {
		// Check if we have a complete line (ends with newline)
		if dbgMsg[msgSize-1] == '\n' {
			debugmsg := strings.TrimSpace(string(dbgMsg[:msgSize]))
			logger.Info("TEST OUTPUT: %s", debugmsg)

			msgSize = 0 // Reset msgSize after printing

			// Check for common test failure indicators
			if strings.Contains(debugmsg, "Failed") || strings.Contains(debugmsg, "FAILED") ||
				strings.Contains(debugmsg, "Error") || strings.Contains(debugmsg, "ERROR") {
				logger.Info("*** TEST FAILED ***")
				return false
			}

			// Check for success indicators
			if strings.Contains(debugmsg, "Passed") || strings.Contains(debugmsg, "PASSED") {
				logger.Info("*** TEST PASSED ***")
			}
		}
	}
	return true
}
