package tests

import (
	"fmt"
	"log"
	"os"
	gameboypackage "pajalic.go.emulator/packages/cpu"
	"testing"
)

func TestCpuWithTestData(t *testing.T) {
	dirPath := "GameboyCPUTests/v2"
	files := getAllTestFiles(dirPath)

	for _, file := range files {
		fmt.Println("FileName:", file.Name())

		LoadJsonTestData(dirPath, file.Name())
		for _, entry := range Testdata {

			var canRun = true

			for _, pair := range entry.Initial.Ram {
				if pair[0] >= 0xE000 {
					canRun = false
				}
			}

			if canRun {

				fmt.Println("==========================================================")
				fmt.Println("Name:", entry.Name)
				fmt.Println("Initial State PC:", entry.Initial.PC)
				fmt.Println("Final State PC:", entry.Final.PC)

				initialState := entry.Initial

				registers := gameboypackage.CpuRegisters{
					initialState.A,
					initialState.F,
					initialState.B,
					initialState.C,
					initialState.D,
					initialState.E,
					initialState.H,
					initialState.L,
					initialState.PC - 1,
					initialState.SP,
				}
				context := gameboypackage.CpuContext{
					Regs:             registers,
					FetchedData:      0,
					MemDest:          0,
					DestIsMem:        false,
					CurOpCode:        0,
					Halted:           false,
					Stepping:         false,
					IntMasterEnabled: false,
					IERegister:       0,
					IntFlags:         0,
				}

				gameboypackage.NewCpu(context)
				gameboypackage.PpuInit()
				gameboypackage.TimerInit()
				gameboypackage.InitInstructions()

				gameboypackage.GetEmuContext().Running = true
				gameboypackage.GetEmuContext().Paused = false
				gameboypackage.GetEmuContext().Ticks = 0

				gameboypackage.ProgramLoad(initialState.Ram)
				//for i, _ := range entry.Cycles {

				gameboypackage.CpuStep()
				fmt.Println("Cycle:", entry.Cycles[0])
				fmt.Printf("PC: %d, Fetch: %d\n", gameboypackage.CpuCtx.Regs.Pc, gameboypackage.CpuCtx.FetchedData)

				//}
				finalState := entry.Final
				ctx := gameboypackage.CpuCtx

				if finalState.A != ctx.Regs.A {
					t.Fatalf("got %d, wanted %d", ctx.Regs.A, finalState.A)
				}
				if finalState.F != ctx.Regs.F {
					t.Fatalf("got %d, wanted %d", ctx.Regs.F, finalState.F)
				}
				if finalState.B != ctx.Regs.B {
					t.Fatalf("got %d, wanted %d", ctx.Regs.B, finalState.B)
				}
				if finalState.C != ctx.Regs.C {
					t.Fatalf("got %d, wanted %d", ctx.Regs.C, finalState.C)
				}
				if finalState.D != ctx.Regs.D {
					t.Fatalf("got %d, wanted %d", ctx.Regs.D, finalState.D)
				}
				if finalState.E != ctx.Regs.E {
					t.Fatalf("got %d, wanted %d", ctx.Regs.E, finalState.E)
				}
				if finalState.H != ctx.Regs.H {
					t.Fatalf("got %d, wanted %d", ctx.Regs.H, finalState.H)
				}
				if finalState.L != ctx.Regs.L {
					t.Fatalf("got %d, wanted %d", ctx.Regs.L, finalState.L)
				}
			}
		}
	}
}

func TestCpuWithSpecificFile(t *testing.T) {
	dirPath := "GameboyCPUTests/v2"

	LoadJsonTestData(dirPath, "f8.json")

	for _, entry := range Testdata {
		if entry.Name == "f8 be b7" {

			fmt.Println("Name:", entry.Name)
			fmt.Println("Initial State PC:", entry.Initial.PC)
			fmt.Println("Final State PC:", entry.Final.PC)

			initialState := entry.Initial

			registers := gameboypackage.CpuRegisters{
				initialState.A,
				initialState.F,
				initialState.B,
				initialState.C,
				initialState.D,
				initialState.E,
				initialState.H,
				initialState.L,
				initialState.PC - 1,
				initialState.SP,
			}
			context := gameboypackage.CpuContext{
				Regs:             registers,
				FetchedData:      0,
				MemDest:          0,
				DestIsMem:        false,
				CurOpCode:        0,
				Halted:           false,
				Stepping:         false,
				IntMasterEnabled: false,
				IERegister:       0,
				IntFlags:         0,
			}

			gameboypackage.NewCpu(context)
			gameboypackage.PpuInit()
			gameboypackage.TimerInit()
			gameboypackage.InitInstructions()

			gameboypackage.GetEmuContext().Running = true
			gameboypackage.GetEmuContext().Paused = false
			gameboypackage.GetEmuContext().Ticks = 0

			gameboypackage.ProgramLoad(initialState.Ram)
			//for i, _ := range entry.Cycles {

			gameboypackage.CpuStep()
			fmt.Println("Cycle:", entry.Cycles[0])
			fmt.Printf("PC: %d, Fetch: %d\n", gameboypackage.CpuCtx.Regs.Pc, gameboypackage.CpuCtx.FetchedData)

			//}
			finalState := entry.Final
			ctx := gameboypackage.CpuCtx

			if finalState.A != ctx.Regs.A {
				t.Fatalf("got %d, wanted %d", ctx.Regs.A, finalState.A)
			}
			if finalState.F != ctx.Regs.F {
				t.Fatalf("got %d, wanted %d", ctx.Regs.F, finalState.F)
			}
			if finalState.B != ctx.Regs.B {
				t.Fatalf("got %d, wanted %d", ctx.Regs.B, finalState.B)
			}
			if finalState.C != ctx.Regs.C {
				t.Fatalf("got %d, wanted %d", ctx.Regs.C, finalState.C)
			}
			if finalState.D != ctx.Regs.D {
				t.Fatalf("got %d, wanted %d", ctx.Regs.D, finalState.D)
			}
			if finalState.E != ctx.Regs.E {
				t.Fatalf("got %d, wanted %d", ctx.Regs.E, finalState.E)
			}
			if finalState.H != ctx.Regs.H {
				t.Fatalf("got %d, wanted %d", ctx.Regs.H, finalState.H)
			}
			if finalState.L != ctx.Regs.L {
				t.Fatalf("got %d, wanted %d", ctx.Regs.L, finalState.L)
			}

		}
	}

}

func getAllTestFiles(dirPath string) []os.FileInfo {
	// Specify the directory path
	// replace with your relative directory path

	// Open the directory
	dir, err := os.Open(dirPath)
	if err != nil {
		log.Fatal(err)
	}
	defer dir.Close()

	// Read the directory entries
	files, err := dir.Readdir(-1)
	if err != nil {
		log.Fatal(err)
	}

	return files
}
