//go:build testrunner && (!js || !wasm)

package main

import (
	"app/internal/common"
	"app/internal/cpu"
	"app/internal/logger"
	"app/internal/memory"
	"app/internal/ui"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

// Test runner for headless GB test ROM execution
func main() {
	var romFile = flag.String("rom", "", "ROM file to test")
	var timeout = flag.Int("timeout", 60, "Timeout in seconds")
	var expectedOutput = flag.String("expect", "Passed", "Expected output string")
	flag.Parse()

	if *romFile == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s -rom <rom_file> [-timeout <seconds>] [-expect <string>]\n", os.Args[0])
		os.Exit(1)
	}

	// Disable boot animation for faster tests
	common.SetBootAnimationEnabled(false)
	common.GlobalConfig.LogLevel = "ERROR" // Quiet mode

	// Start emulator
	logger.Info("Starting test ROM: %s", *romFile)
	emuInstance := ui.StartEmulator(*romFile)
	if emuInstance == nil {
		fmt.Fprintf(os.Stderr, "Failed to start emulator\n")
		os.Exit(1)
	}

	// Run headless
	startTime := time.Now()
	timeoutDuration := time.Duration(*timeout) * time.Second
	var serialOutput strings.Builder
	lastFrameTime := time.Now()

	for {
		// Check timeout
		if time.Since(startTime) > timeoutDuration {
			fmt.Fprintf(os.Stderr, "TIMEOUT: Test did not complete in %d seconds\n", *timeout)
			fmt.Fprintf(os.Stderr, "Serial output:\n%s\n", serialOutput.String())
			os.Exit(1)
		}

		// Run one frame of emulation
		emuInstance.StepFrame()

		// Check serial output after each frame
		now := time.Now()
		if now.Sub(lastFrameTime) > 16*time.Millisecond {
			lastFrameTime = now
			
			// Collect all serial data
			for {
				if memory.BusCtx().BusRead(0xFF02) == 0x81 {
					c := memory.BusCtx().BusRead(0xFF01)
					if c >= 32 && c <= 126 || c == '\n' || c == '\r' {
						serialOutput.WriteByte(c)
					}
					memory.BusCtx().BusWrite(0xFF02, 0)
				} else {
					break
				}
			}

			output := serialOutput.String()
			
			// Check for completion indicators
			if strings.Contains(output, *expectedOutput) {
				fmt.Println("✅ TEST PASSED")
				fmt.Println("\nSerial output:")
				fmt.Println(output)
				os.Exit(0)
			}
			
			if strings.Contains(output, "Failed") {
				fmt.Fprintf(os.Stderr, "❌ TEST FAILED\n")
				fmt.Fprintf(os.Stderr, "\nSerial output:\n%s\n", output)
				os.Exit(1)
			}
		}

		time.Sleep(1 * time.Millisecond)
	}
}
