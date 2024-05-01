; This program acts as a simple kernel, built off of the CPU implementation
; in the go files. The goal of the kernel is to load a user program into memory,
; and bound the user program so it cannot load or store from outside its allotted memory or
; execute privileged instructions without using a syscall.
;
; The kernel implementation is split in to two parts:
;
;1. Bootloader
;	The kernel starts by calling setTrapHandler, which tells the CPU where
;	to jump to if a trap is detected.
;	Then, the kernel loads the user program into memory, starting at memory address 1024.
;	Finally, the kernel calls changeMode to switch into user mode, and jumps to the start
;	of the user program.
;2. TrapHandler
;	The trapHandler is made to handle syscalls, illegal instruction calls, 
;	and memory out of bounds errors. The trapHandler starts by storing r0-r5 in memory
;	(r6 is not touched, and r7 is preserved in the go code), then loads memory address 6
;	into r5. This is the trap number, which is stored in the Go code kernelTrap call.
;	If the trap number < 3, it is a syscall, so the appropriate syscall is executed, then the
;	CPU registers are restored, and r7 is returned to where the program left off.
;	If the trap number == 4, it is a memory out of bounds issue, so the appropriate
;	message is written to the output device, and the program is halted.
;	If the trap number == 5, it is an illegal instruction issue, so the appropriate
;	message is written to the output device, and the program is halted.
;	If the trap number ==6, the timer has fired, so the appropriate message is written to
;	the output device, and the program resumes

main:
	setTrapHandler .trapHandler

; Store first two bytes of input into r0
bootloader:
	read r0							; read 1/2 bytes
	shl r0 8 r0						; shift left 8bits
	read r1							; read 2/2 bytes
	or r0 r1 r0

	loadLiteral 1024 r2				; r2: 1024, just preserves the amount we need to buffer the program by

	loadLiteral 0 r3				; make r3 to 0

; Next: Keep reading from input into r1 until we get 0
loop:
	; First: Read a word (4 bytes) into r1, r2, r3, and r4
	loadLiteral 0 r4				; make r4 to 0
	
	read r1							; read r1 because it'll overwrite
	shl r1 24 r1					; shift left r1 to 24 and store that into r1
	or r4 r1 r4						; r4 || r1 and store that into r4
	
	read r1							; read r1
	shl r1 16 r1					; shift r1 left 16 
	or r4 r1 r4						; combine r1 into r4

	read r1							; repeat
	shl r1 8 r1
	or r4 r1 r4
	
	read r1							; no need to shift this time
	or r4 r1 r4

	store r4 r2						; store the word into memory address

	load r2 r4						; load r2 to r4
	
	add r2 1 r2						; add 1 to r2
	add r3 1 r3						; add 1 to r3

	lt r3 r0 r4		; Check if r4, our word counter, is equal to r0, the program length
	
	cmove r4 .loop r7

	changeMode 0
	loadLiteral 1024 r7		; Move the iptr to 1024, where the program starts

trapHandler:
	; Store the CPU program state in memory before executing trap handler

	store r0 0
	store r1 1
	store r2 2
	store r3 3
	store r4 4
	store r5 5

	load 6 r5
	loadLiteral .exitTrapHandler r4
	
	lt r5 3 r0						; If r5 < 3, the trap is a syscall
	cmove r0 .syscallHandler r7

	eq r5 4 r0						; If r5 == 4, the trap is a memory bounds issue
	cmove r0 .memoryOutOfBounds r7

	eq r5 5 r0						; If r5 == 5, the trap is an illegal instruction issue
	cmove r0 .illegalInstruction r7

	eq r5 6 r0						; If r5 == 6, a timer has fired
	loadLiteral .timerFired r3
	cmove r0 r3 r7

	eq r5 7 r0						; if r5 == 7, we are writing the timer fire count
	loadLiteral .writeTimerCount r3
	cmove r0 r3 r7

	move r4 r7		; If r5 is none of those, it is an unrecognized trap, so exit
	
syscallHandler:
	; If r6 == 0, execute a read instruction
	; If r6 == 1, execute a write instruction
	; If r6 == 2, execute a halt instruction

	eq r5 0 r0
	loadLiteral .readInstr r3
	cmove r0 r3 r7
	
	eq r5 1 r0
	loadLiteral .writeInstr r3
	cmove r0 r3 r7
	
	eq r5 2 r0
	loadLiteral .haltInstr r3
	cmove r0 r3 r7

	; exitTrapHandler here, just in case r6 is somehow < 0
	move r4 r7

readInstr:
	read r6
	move r4 r7

writeInstr:
	write r6
	move r4 r7
	
haltInstr:
	write 10
	write 'P'
	write 'r'
	write 'o'
	write 'g'
	write 'r'
	write 'a'
	write 'm'
	write 32
	write 'h'
	write 'a'
	write 's'
	write 32
	write 'e'
	write 'x'
	write 'i'
	write 't'
	write 'e'
	write 'd'
	write 10
	halt

memoryOutOfBounds:
	write 10
	write 'O'
	write 'u'
	write 't'
	write 32
	write 'o'
	write 'f'
	write 32
	write 'b'
	write 'o'
	write 'u'
	write 'n'
	write 'd'
	write 's'
	write 32
	write 'm'
	write 'e'
	write 'm'
	write 'o'
	write 'r'
	write 'y'
	write 32
	write 'a'
	write 'c'
	write 'c'
	write 'e'
	write 's'
	write 's'
	write '!'
	write 10
	halt
	
illegalInstruction:
	write 10
	write 'I'
	write 'l'
	write 'l'
	write 'e'
	write 'g'
	write 'a'
	write 'l'
	write 32
	write 'i'
	write 'n'
	write 's'
	write 't'
	write 'r'
	write 'u'
	write 'c'
	write 't'
	write 'i'
	write 'o'
	write 'n'
	write '!'
	write 10
	halt

timerFired:
	write 10
	write 'T'
	write 'i'
	write 'm'
	write 'e'
	write 'r'
	write 32
	write 'f'
	write 'i'
	write 'r'
	write 'e'
	write 'd'
	write '!'
	write 10
	write 48
	move r4 r7

writeTimerCount:
	write 'T'
	write 'i'
	write 'm'
	write 'e'
	write 'r'
	write 32
	write 'f'
	write 'i'
	write 'r'
	write 'e'
	write 'd'
	write 32

	load 8 r0				; Load the timerFiredCount from memory
	loadLiteral 28 r1		; Shift amount: This will decrease by 4 each iteration

timerLoop:
	shr r0 r1 r2			; Get leftmost un-written four bits
	and r2 15 r2			; Mask the leftmost un-written four bits
	lt r2 10 r3				; Check: Are those four bits less than 10?

	loadLiteral .numeric r5
	cmove r3 r5 r7			; If r2 is less than 10, jump to numeric
	add r2 55 r2			; If r2 is greater than 10, add 55 so it becomes the proper ASCII for A-F

continue:
	write r2				; Write r2
	sub r1 4 r1				; Reduce the shift amount by 4

	lt r1 0 r3				; Check: Is the shift amount LESS than 0?
	loadLiteral .finishTimerCount r5
	cmove r3 r5 r7			; If so, jump to the rest of the writeTimerCount
	loadLiteral .timerLoop r5
	move r5 r7				; Otherwise, do the loop again

numeric:
	add r2 48 r2			; Add 48 to r2 so it becomes the proper ascii for 0-9
	loadLiteral .continue r5
	move r5 r7				; Jump back to continue the loop
	
finishTimerCount:
	write 32
	write 't'
	write 'i'
	write 'm'
	write 'e'
	write 's'
	write 10
	halt
	

exitTrapHandler:
	; When exiting the trapHandler: Restore all register states
	load 0 r0
	load 1 r1
	load 2 r2
	load 3 r3
	load 4 r4
	load 5 r5

	; Then: Restore the iptr back to where it was in the program.
	; In kernel.go, we stored the correct iptr value at memory location 7.

	changeMode 0 7
