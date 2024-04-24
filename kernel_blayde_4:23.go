package main

import "fmt"

// ************* Kernel support *************
//
// All of your CPU emulator changes for Assignment 2 will go in this file.

// The state kept by the CPU in order to implement kernel support.
type kernelCpuState struct {
	// TODO: Fill this in.
	syscall_id uint8  // only need 1-3
    kernelMode bool   // false for userland mode, true for kernel mode
    timer uint32 // Instruction count for timer interrupts
	interrupt bool // allowed to interrupt?
	trap_handler uint32 // " jump to the same address in kernel memory"
}

// The initial kernel state when the CPU boots.
var initKernelCpuState = kernelCpuState{
	// TODO: Fill this in.
	// iptr: 0,
    // mode: false, // Start in user mode
    // timer: 0,
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

	// TODO: Add hooks to other existing instructions to implement kernel
	// support.

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
