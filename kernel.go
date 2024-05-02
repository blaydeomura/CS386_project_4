package main

import "fmt"

// ************* Kernel support *************
//
// All of your CPU emulator changes for Assignment 2 will go in this file.

/*
CPU support for kernel implementation.

Contents:
* 1.x: CPU support setup: Struct + Var init
* 2.x: Helper functions
* 3.x: Pre-execute hook parts
* 4.x: Instrucitons
* 5.x: Syscall
* 6.x: Instruction hooks

1.1: Setup:
* Purpose:
	* The purpose of the setup is to save the state of the CPU to implement kernel support
	* This will be used later for functions, pre execute hook, instructions, and instruction hooks
* kernelCpuState struct
	* Features include:
		* kernelMode: Tells kernel whether in kernel mode or userland mode
		* timerCount: Keeps track of the timer slice count and resets every 128 counts
		* trapHandlerAddr: This is the memory address where our kernel mode lives
		* timerFireCount: This is how many times the timerCount has hit 128
* 1.2: var initKernelCpuState
	* This sets up the initial state for the cpu
	* kernelMode: we initially set the kernelMode to true so we start in kernel mode on
	* timerCount: starts at 0 and is to be incremented
	* trapHandlerAddr: it starts as 0 for now but will be set once we get access to registers and memory
	* timerFireCount: starts at 0 and increments every 128 time slices

2.x Helper Functions
* 2.1: kernelTrap
* This function is used to trap every time the mode needs to be switched into kernel mode
* What it does:
	* Saves the timerFireCount into kernel.asm memory 
	* Saves the trapNumber(parameter that is called in asm) into kernel.asm memory
	* Saves the instruction pointer into the kernel.asm memory
	* Changes the kernel mode to true
	* Sets the instruction pointer to the allocated memory for kernel mode at trapHandlerAddr
* Where it is used:
	* preExecuteHook: In the preExecuteHook, if the timerCount is greater than 128, then we need to trap to the kernel mode
	* Instruction hooks: when instruction hooks fail, this function will be called to trap into kernel mode
	* Instructions: It is also used in instructions for kernel.asm to use for setting the trap handler as well as changing mode
	* Syscall: Also always called in syscall to trap to kernel mode


3.x preExecuteHooks
* Purpose:  A hook which is executed at the beginning of each instruction step.
* 3.1: Check timer count
	* We first check if the timer count is greater than 128
		* If it is greater than 128 then we call kernelTrap, increment both timerFireCount and timer coiunt
* 3.2: Check if kernel mode
	* If we're not in kernel mode, then we increment counter

4.x Instructions
* Purpose: provide callable instructions for kernel.asm
* 4.1 instrSetTrapHandler
	* Purpose: setTrapHandler takes one argument, which should be the trapHandler address in kernel.asm.
		* setTrapHandler sets the kernel trapHandlerAddr to args[0]
	* Process:
		* Identifies the setTrapHandler name 
		* sets trap handler address
		* runs the genValidate function from instr.go
* 4.2 instrChangeMode
	* Purpose: It is meant to change kernel modes and to switch back into the user program at the same time.
	* Process:
		* instrChangeMode checks first if the kernel is in kernelMode or not. If not, it does nothing
		* If it is called in kernel mode, instrChangeMode changes the kernelMode to parameter 1, which should be either
		* 0 or 1.
		* 0: Switch to user mode, so kernelMode = false
		* 1: Switch to kernel mode, so kernelMode = true
		* Then, if the second argument passed to instrChangeMode is not empty, it loads the value at args[1] and 
		* sets the iptr to that value

5.x Syscall
* Purpose: to allow a way for userland process to request services from OS kernel
* Process:
	* Mask out the high bit to get the correct syscall number
	* Check whether it is a valid syscall so either 0, 1, or 2
	* Then call kernelTrap

6.x Instruction Hooks
* Purpose: Purpose is to provide specific behavior for instructions
	* good for preventing execution of privileged instructions when in user mode
* 6.1: instrWrite:
	* Purpose: instruction hook check for write
	* if not in kernel mode then call kernelTrap() and return true
	* else return false
* 6.2: instrRead:
	* Purpose: instruction hook check for read
	* if not in kernel mode then call kernelTrap() and return true
	* else return false
* 6.3: instrUnreachable:
	* Purpose: Hook to try and execute unreachable
	* if not in kernel mode then call kernelTrap() and return true
	* else return false
* 6.4: instrLoad:
	* Purpose: Hook to check if loading from valid memory bounds
	* if not in kernel mode
	* Set the address to argument 0
	* Ensure addr is in 1024 - 2048
* 6.5: instrStore:
	* Purpose: Hook to check if store from valid memory bounds
	* if not in kernel mode
	* Set the address to argument 0
	* Ensure addr is in 1024 - 2048
* 6.6: instrHalt:
	* Purpose: Hook to halt cpu
	* if not in kernel mode then call kernelTrap() and return true
	* else return false
* 6.7: instrSetTrapHandler:
	* Purpose: is to set trap handler
	* if not in kernel mode then call kernelTrap() and return true
	* else return false
* 6.8: instrChangeMode:
	* Purpose: Hook to check if changing mode is allwoed
	* if not in kernel mode then call kernelTrap() and return true
	* else return false
*/

// The state kept by the CPU in order to implement kernel support.
type kernelCpuState struct {
	kernelMode      bool   // True if in kernel mode, false if in user mode.
	timerCount      uint32 // Count of instructions executed for timer management.
	trapHandlerAddr word // Static memory address where the trap handler is located.
	timerFireCount	uint32 // how many time the timer fires
}

// The initial kernel state when the CPU boots.
var initKernelCpuState = kernelCpuState{
	kernelMode:      true,  // Start in kernel mode.
	timerCount:      0,      // Timer count starts at 0.
	trapHandlerAddr: 0, 	// Kernel Addr
	timerFireCount:	 0,		// how many times the timer fires
}

// This is trap for kernel
func kernelTrap(c *cpu, trapNumber word) {
	c.memory[8] = word(c.kernel.timerFireCount)
	c.memory[6] = trapNumber
	c.memory[7] = c.registers[7] // save iptr in memory
	c.kernel.kernelMode = true	// switch to kernel mode
	c.registers[7] = c.kernel.trapHandlerAddr
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
	// Check for timer interrupt every 128 instructions.
	if k.timerCount > 128 {
		kernelTrap(c, 6)
		k.timerFireCount += 1
		k.timerCount = 0
	}

	// if not in kernel mode increment the timer
	if !c.kernel.kernelMode {
		k.timerCount++;
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

	// Hook for write instruction to check for privellages
	instrWrite.addHook(func(c *cpu, args [3]byte) (bool, error) {
		if !c.kernel.kernelMode {
			kernelTrap(c, 5)
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
			return true, nil
		}
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
