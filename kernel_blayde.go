package main

import "fmt"

// ************* Kernel support *************
//
// All of your CPU emulator changes for Assignment 2 will go in this file.

// The state kept by the CPU in order to implement kernel support.
type kernelCpuState struct {
	// TODO: Fill this in.
    kernelMode bool       // True if in kernel mode, false if in user mode.
    timerCount uint32     // Count of instructions executed for timer management.
    trapHandlerAddr uint32 // Static memory address where the trap handler is located.
}

// The initial kernel state when the CPU boots.
var initKernelCpuState = kernelCpuState{
	// TODO: Fill this in.
    kernelMode: false,   // Start in user mode.
    timerCount: 0,       // Timer count starts at 0.
    trapHandlerAddr: 0x1000, // NEED TO PUT ADDRESS OF KERNEL MODE HERE?
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
    // Increment the instruction timer on each CPU step.
    k.timerCount++
    
	// Security check 1:
    // Check for timer interrupt every 128 instructions.
    if k.timerCount % 128 == 0 {
        fmt.Println("\nTimer fired!\n")
        // TODO: Need to halt and interrupt here? need to switch to kernel mode?
    }

	// Secruity check 2:
    // Check for privileged instructions if in user mode.
    if !k.kernelMode {
        currentInstr := c.memory[c.registers[7]] // Assuming IP is in register 7.
        if currentInstr.IsPrivileged() {
            return false, fmt.Errorf("Illegal instruction access in user mode")
        }
    }

	// ADD MORE Checks or state update below-------

	// TODO: If syscall is called, make sure in kernel mode

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

	// TODO: Ensure that 'load' can only access memory within bounds
    instrLoad.addHook(func(c *cpu, args [3]byte) (bool, error) {
        addr := resolveArg(c, args[0])
        if addr < 1024 || addr >= 2048 {
            return true, fmt.Errorf("\nOut of bounds memory access!\n")
        }
        return false, nil
    })
	

	// TODO: Ensure that 'store' can only access memory within bounds
	instrStore.addHook(func(c *cpu, args [3]byte) (bool, error) {
        addr := resolveArg(c, args[1])
        if addr < 1024 || addr >= 2048 {
            return true, fmt.Errorf("\nOut of bounds memory access!\n")
        }
        return false, nil
    })

	// TODO: Privellaged instructions like 'halt', 'read', 'write' should only be executed in kernel
	// For syscall essentially
	instrSyscall.addHook(func(c *cpu, args [3]byte) (bool, error) {
        if !c.kernel.kernelMode {
            // Switch to kernel mode and set the trap handler address
            c.kernel.kernelMode = true
            c.kernel.trapHandlerAddr = c.registers[7]
            c.registers[7] = k.trapHandlerAddr // Jump to the trap handler
            return true, nil // Skip the current instruction execution as we handle the mode switch
        }
        return false, nil
    })


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
				syscallNumber := int(args[0] & 0x7F)  // Mask out the high bit to get the correct syscall number
				fmt.Println("Executing syscall number: ", syscallNumber)
				
				// TODO: ADD IN a check to see if we're in kernel mode. If not, get into it


				switch syscallNumber {  
				case 0: // read syscall
					var buf [1]byte
					_, err := c.read.Read(buf[:])
					if err != nil {
						return fmt.Errorf("read error: %v", err)
					}
					c.registers[6] = word(buf[0]) // Store the read byte into register 6
					return nil
		
				case 1: // write syscall
					b := byte(c.registers[6] & 0xFF)    // Get the least significant byte from register 6
					_, err := c.write.Write([]byte{b})
					if err != nil {
						return fmt.Errorf("write error: %v", err)
					}
					return nil
		
				case 2: // exit syscall
					fmt.Println("Program has exited")
					c.halted = true // Halt the CPU
					return nil
		
				default:
					return fmt.Errorf("unknown syscall number %d", syscallNumber) // Use the masked number in error messages too
				}
			},
			validate: func(args [3]byte) error {
				syscallNumber := int(args[0] & 0x7F)  // Mask out the high bit
				if syscallNumber > 2 {
					return fmt.Errorf("invalid syscall number %d", syscallNumber)
				}
				return nil
			},
		}

		// TODO: Add other instructions that can be used to implement a kernel.
	)

	// Add kernel instructions to the instruction set.
	instructionSet.add(instrSyscall)
}
