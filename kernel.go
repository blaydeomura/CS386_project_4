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
}

// The initial kernel state when the CPU boots.
var initKernelCpuState = kernelCpuState{
	// TODO: Fill this in.
	kernelMode:      true,  // Start in kernel mode.
	timerCount:      0,      // Timer count starts at 0.
	trapHandlerAddr: 0, // TODO: NEED TO PUT ADDRESS OF KERNEL MODE HERE?
}

// -----For preexecute hook, this is where all of the security concerns are done----
// -----ie making sure enough space allocation, others listed on the document-------
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
	// TODO: Fill this in.
	// TODO: If syscall is called, make sure in kernel mode
    // if is a syscall // Assuming isSyscall checks if current instruction is a syscall
    //     iptr registers[7] = k.trapHandlerAddr // Set instruction pointer to trap handler address
    //     return true, nil // Skip normal execution to handle the syscall in kernel mode
    // }


	// Increment the instruction timer on each CPU step.
	//k.timerCount++

	// Security check 2:
	// Check for timer interrupt every 128 instructions.
	//if k.timerCount % 128 == 0 {
	//	fmt.Println("\nTimer fired!\n")
		// TODO: Need to halt and interrupt here? need to switch to kernel mode?
	//}

	// Secruity check 3:
	// Check for privileged instructions if in user mode.
	//if !k.kernelMode {
		// currentInstr := c.memory[c.registers[7]] // Assuming IP is in register 7.
		//TODO: check if this instruction has privilleges
		// if currentInstr.IsPrivileged() {
		// 	return false, fmt.Errorf("Illegal instruction access in user mode")
		// }
	//}

	// ADD MORE Checks or state update below-------

	// TODO: Manage the behavior when this pre execute hook fails

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

	// TODO: Add hooks to other existing instructions to implement kernel support.

	//TODO: Hook for write instruction to check for privellages
	instrWrite.addHook(func(c *cpu, args [3]byte) (bool, error) {
		if !c.kernel.kernelMode {
			return true, fmt.Errorf("write operation not allowed in user mode")
		}
		return false, nil
	})

	// TODO: If called in userland, ensure that 'load' can only access memory within bounds
	instrLoad.addHook(func(c *cpu, args [3]byte) (bool, error) {
		if !c.kernel.kernelMode {
			addr := resolveArg(c, args[0])
			if addr < 1024 || addr >= 2048 {
				return true, fmt.Errorf("\nOut of bounds memory access!\n")
			}
		}
		return false, nil
	})

	// TODO: If called in userland, ensure that 'store' can only access memory within bounds
	instrStore.addHook(func(c *cpu, args [3]byte) (bool, error) {
		if !c.kernel.kernelMode {
			addr := resolveArg(c, args[1])
			if addr < 1024 || addr >= 2048 {
				return true, fmt.Errorf("\nOut of bounds memory access!\n")
			}
		}
		return false, nil
	})

	//TODO: implement hook for read

	//TODO: implement hook for halt

	// TODO: Privellaged instructions like 'halt', 'read', 'write' should only be executed in kernel
	// For syscall essentially
	// instrSyscallCheck.addHook(func(c *cpu, args [3]byte) (bool, error) {
	// 	if !c.kernel.kernelMode {
	// 		// Switch to kernel mode and set the trap handler address
	// 		c.kernel.kernelMode = true
	// 		//TODO: Get register 7 iptr state
	// 		//TODO: GET TRAP HANDLER ADDRESS
	// 		return true, nil                   // Skip the current instruction execution as we handle the mode switch
	// 	}
	// 	return false, nil
	// })

	//------ Instructions below ----

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

				c.memory[24] = word(syscallNumber)
				c.memory[28] = c.registers[7]
				c.registers[7] = c.kernel.trapHandlerAddr
				c.kernel.kernelMode = true
				//fmt.Println("Moved into trap handler...")
				return nil
			},
			validate: genValidate(regOrLit, ignore, ignore),

			/*func(args [3]byte) error {
				if syscallNumber > 2 {
					return fmt.Errorf("invalid syscall number %d", syscallNumber)
				}
				return nil 
			} */
		}

		// TODO: Add other instructions that can be used to implement a kernel.

		// setTrapHandler takes one argument, which should be the trapHandler address in kernel.asm.
		// setTrapHandler sets the kernel trapHandlerAddr to args[0]
		instrSetTrapHandler = &instr{
			name: "setTrapHandler",
			cb: func(c *cpu, args [3]byte) error {
				fmt.Print("Hello")
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

	// Add kernel instructions to the instruction set.
	instructionSet.add(instrSyscall)
	instructionSet.add(instrSetTrapHandler)
	instructionSet.add(instrChangeMode)
}
