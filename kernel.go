package main

import "fmt"

// ************* Kernel support *************
//
// All of your CPU emulator changes for Assignment 2 will go in this file.

// The state kept by the CPU in order to implement kernel support.
type kernelCpuState struct {
	// TODO: Fill this in.
	kernelMode      bool   // True if in kernel mode, false if in user mode.
	timerCount      uint32 // Count of instructions executed for timer management.
	trapHandlerAddr word // Static memory address where the trap handler is located.
	timerFireCount	uint32 // how many time the timer fires
}

// The initial kernel state when the CPU boots.
var initKernelCpuState = kernelCpuState{
	// TODO: Fill this in.
	kernelMode:      true,  // Start in kernel mode.
	timerCount:      0,      // Timer count starts at 0.
	trapHandlerAddr: 0, 	// Kernel Addr
	timerFireCount:	 0,		// how many times the timer fires
}

// This is trap for kernel
func kernelTrap(c *cpu, trapNumber word) {
	// if word = 6, then load c,memory[7] = timerFiredCount
	c.memory[8] = word(c.kernel.timerFireCount)

	c.memory[6] = trapNumber
	c.memory[7] = c.registers[7] // save iptr in memory
	c.kernel.kernelMode = true	// switch to kernel mode
	c.registers[7] = c.kernel.trapHandlerAddr
}

// saveCpuState saves the current state of the CPU registers into a designated area of memory.
func saveCpuState(c *cpu) {
    fmt.Println("Saving current CPU state...")
    // Save registers to a predefined memory area, e.g., starting at address 0.
    for i, reg := range c.registers {
        c.memory[4*i] = reg // Assume each register value is stored in 4 consecutive bytes.
    }
    fmt.Println("CPU state saved successfully.")
}

// A hook which is executed at the beginning of each instruction step.
//
// This permits the kernel support subsystem to perform extra validation that is
// not part of the core CPU emulator functionality.
//
// If `preExecuteHook` returns an error, the CPU is considered to have entered
// an illegal state, and it halts.
//
// If `preExecuteHook` returns `true`, the instruction is "skipped": `cpu.step`
// will immediately return without any further execution.
func (k *kernelCpuState) preExecuteHook(c *cpu) (bool, error) {
	// if not in kernel mode increment the timer
	if !c.kernel.kernelMode {
		k.timerCount++;
	}

	// Check for timer interrupt every 128 instructions.
	if k.timerCount > 128 {
		kernelTrap(c, 6)
		k.timerFireCount += 1
		k.timerCount = 0
		//fmt.Println("\nTimer fired!")
	}

	return false, nil
}

// Initialize kernel support.
//
// (In Go, any function named `init` automatically runs before `main`.)
func init() {
	if false {
		// This is an example of adding a hook to an instruction. You probably
		// don't actually want to add a hook to the `add` instruction.
		instrAdd.addHook(func(c *cpu, args [3]uint8) (bool, error) {
			a0 := resolveArg(c, args[0])
			a1 := resolveArg(c, args[1])
			if a0 == a1 {
				// Adding a number to itself? That seems like a weird thing to
				// do. Best just to skip it...
				return true, nil
			}

			if args[2] == 7 {
				// This instruction is trying to write to the instruction
				// pointer. That sounds dangerous!
				return false, fmt.Errorf("You're not allowed to ever change the instruction pointer. No loops for you!")
			}

			return false, nil
		})
	}

	// Basic hook looks  like this:
		// if a hook is not supposed to happen, then we need to switch to kernel mode
		// else if it is not called, then we're good to continue

	// Hook for write instruction to check for privellages
	instrWrite.addHook(func(c *cpu, args [3]byte) (bool, error) {
		if !c.kernel.kernelMode {
			kernelTrap(c, 5)
			// instrHalt.cb(c, args)
			return true, nil
		}
		return false, nil
	})

	// Hook to try and Read
	instrRead.addHook(func(c *cpu, args [3]byte) (bool, error) {
		if !c.kernel.kernelMode {
			kernelTrap(c, 5)
			return true, nil
		}
		return false, nil
	})

	// Hook to try and execute unreachable
	instrUnreachable.addHook(func(c *cpu, args [3]byte) (bool, error) {
	if !c.kernel.kernelMode {
		kernelTrap(c, 5)
		return true, nil
	}
	return false, nil
	})

	// If called in userland, ensure that 'load' can only access memory within bounds
	instrLoad.addHook(func(c *cpu, args [3]byte) (bool, error) {
		if !c.kernel.kernelMode {
			addr := resolveArg(c, args[0])
			if addr < 1024 || addr >= 2048 {
				kernelTrap(c, 4)
				// instrHalt.cb(c, args)
				return true, nil
			}
		}
		return false, nil
	})

	// If called in userland, ensure that 'store' can only access memory within bounds
	instrStore.addHook(func(c *cpu, args [3]byte) (bool, error) {
		if !c.kernel.kernelMode {
			addr := resolveArg(c, args[1])
			if addr < 1024 || addr >= 2048 {
				kernelTrap(c, 4)
				// instrHalt.cb(c, args)
				return true, nil
			}
		}
		return false, nil
	})

	// Hook to halt cpu
	instrHalt.addHook(func(c *cpu, args [3]byte) (bool, error) {
		if !c.kernel.kernelMode {
			kernelTrap(c, 5)
			// c.kernel.kernelMode = true
			// instrHalt.cb(c, args)
			return true, nil
		}
		//fmt.Printf("Timer fired %8.8x times\n", c.kernel.timerFireCount)
		//fmt.Printf("Timer fired: %d times\n", c.kernel.timerFireCount)
		//kernelTrap(c, 7)
		return false, nil
	})

	// ---At 4/18, we test by running: go run *.go bootloader.asm prime.asm---
	// Syscall is how a process asks kernel to do something
	var (
		// syscall <code>
		//
		// Executes a syscall. The first argument is a literal which identifies
		// what kernel functionality is requested:
		// - 0/read:  Read a byte from the input device and store it in the
		//            lowest byte of r6 (and set the other bytes of r6 to 0)
		// - 1/write: Write the lowest byte of r6 to the output device
		// - 2/exit:  The program exits; print "Program has exited" and halt the
		// 	 		  machine.
		//
		// You may add new syscall codes if you want, but you may not modify
		// these existing codes, as `prime.asm` assumes that they are supported.
		instrSyscall = &instr{
			name: "syscall",
			cb: func(c *cpu, args [3]byte) error {
				syscallNumber := int(args[0] & 0x7F) // Mask out the high bit to get the correct syscall number
				//fmt.Println("Executing syscall number: ", syscallNumber)

				// TODO: ADD IN a check to see if we're in kernel mode. If not, get into it
				if syscallNumber > 2 || syscallNumber < 0 {
					return fmt.Errorf("invalid syscall number %d", syscallNumber)
				}

				kernelTrap(c, word(syscallNumber))
				return nil
			},
			validate: genValidate(regOrLit, ignore, ignore),
		}

		// setTrapHandler takes one argument, which should be the trapHandler address in kernel.asm.
		// setTrapHandler sets the kernel trapHandlerAddr to args[0]
		instrSetTrapHandler = &instr{
			name: "setTrapHandler",
			cb: func(c *cpu, args [3]byte) error {
				c.kernel.trapHandlerAddr = resolveArg(c, args[0])
				return nil
			},
			validate: genValidate(regOrLit, ignore, ignore),
		}

		// instrChangeMode can take 1 or 2 args. 
		// It is meant to change kernel modes and to switch back into the user program at the same time
		// instrChangeMode checks first if the kernel is in kernelMode or not. If not, it does nothing
		// If it is called in kernel mode, instrChangeMode changes the kernelMode to parameter 1, which should be either
		// 0 or 1.
		// 0: Switch to user mode, so kernelMode = false
		// 1: Switch to kernel mode, so kernelMode = true
		// Then, if the second argument passed to instrChangeMode is not empty, it loads the value at args[1] and 
		// sets the iptr to that value
		instrChangeMode = &instr{
			name: "changeMode",
			cb: func(c *cpu, args [3]byte) error {
				if c.kernel.kernelMode {
					newMode := int(resolveArg(c, args[0]))
					if newMode == 0 {
						c.kernel.kernelMode = false
					} else if newMode == 1 {
						c.kernel.kernelMode = true
					} else {
						return fmt.Errorf("Invalid kernel mode %d", newMode)
					}

					branchAddr := int(args[1] & 0x7F)
					if branchAddr != 0 {
						c.registers[7] = c.memory[branchAddr]
					}
				}
				return nil
			},
			validate: genValidate(regOrLit, regOrLit, ignore),
		}
	)

	instrSetTrapHandler.addHook(func(c *cpu, args [3]byte) (bool, error) {
		if !c.kernel.kernelMode {
			kernelTrap(c, 5)
			return true, nil
		}
		return false, nil
	})

	instrChangeMode.addHook(func(c *cpu, args [3]byte) (bool, error) {
		if !c.kernel.kernelMode {
			kernelTrap(c, 5)
			return true, nil
		}
		return false, nil
	})

	// Add kernel instructions to the instruction set.
	instructionSet.add(instrSyscall)
	instructionSet.add(instrSetTrapHandler)
	instructionSet.add(instrChangeMode)
}
