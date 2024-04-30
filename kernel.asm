; This program acts as a simple kernel, built off of the CPU implementation
; in the go files. The first 

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
	store r1 4
	store r2 8
	store r3 12
	store r4 16
	store r5 20

	load 24 r5
	lt r5 3 r0		; If r6 < 3: r6 == 0, 1, or 2, so it is a syscall
	cmove r0 .syscallHandler r7

	eq r5 4 r0
	cmove r0 .memoryOutOfBounds r7

	eq r5 5 r0
	cmove r0 .illegalInstruction r7

	move .exitTrapHandler r7

syscallHandler:
	; If r6 == 0, execute a read instruction
	; If r6 == 1, execute a write instruction
	; If r6 == 2, execute a halt instruction
	
	eq r5 0 r0
	cmove r0 .readInstr r7
	eq r5 1 r0
	cmove r0 .writeInstr r7
	eq r5 2 r0
	cmove r0 .haltInstr r7

	; Add an exitTrapHandler here, just in case r6 is somehow < 0

readInstr:
	read r6
	move .exitTrapHandler r7

writeInstr:
	write r6
	move .exitTrapHandler r7
	
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

exitTrapHandler:
	; When exiting the trapHandler: Restore all register states
	load 0 r0
	load 4 r1
	load 8 r2
	load 12 r3
	load 16 r4
	load 20 r5

	; Then: Restore the iptr back to where it was in the program.
	; In kernel.go, we stored the correct iptr value at memory location 28.

	changeMode 0 28
