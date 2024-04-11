; Store first two bytes of input into r0
main:
	read r0
	shl r0 8 r0
	read r0

	loadLiteral 1024 r2				; r2: 1024, just preserves the amount we need to buffer the program by
	loadLiteral 0 r3				; r3: Stores how many instructions we have stored in memory so far

; Next: Keep reading from input into r1 until we get 0
loop:
	; First: Read a word (4 bytes) into r1
	loadLiteral 0 r1
	read r1
	shl r1 8 r1
	read r1

	debug 0
	
	shl r1 8 r1
	read r1
	shl r1 8 r1
	read r1

	; Then: 
	add r2 r3 r4	; r4 contains the address where we want to store r1
	store r1 r4		; Store the word instruction into r4

	add r3 1 r3

	eq r1 0 r5
	
	cmove r5 .loop r7

end:
	write 'd'
	write 'o'
	write 'n'
	write 'e'
	write 10
	debug 0
	move r2 r7
