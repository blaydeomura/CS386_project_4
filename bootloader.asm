; Store first two bytes of input into r0
main:
	read r0
	;read r4
	shl r0 8 r0
	;or r0 r4 r0
	read r0

	debug 0

	loadLiteral 1024 r5				; r2: 1024, just preserves the amount we need to buffer the program by

	loadLiteral 0 r2

; Next: Keep reading from input into r1 until we get 0
loop:
	; First: Read a word (4 bytes) into r1, r2, r3, and r4
	loadLiteral 0 r3
	
	read r1
	shl r1 24 r1
	or r3 r1 r3

	;debug 0
	
	read r1
	shl r1 16 r1
	or r3 r1 r3

	;debug 0

	read r1
	shl r1 8 r1
	or r3 r1 r3
	
	;debug 0

	read r1
	store r1 r5
	or r3 r1 r3

	load r5 r3
	debug 0
	
	add r5 1 r5
	add r2 1 r2

	;debug 0

	lt r2 r0 r3		; Check if r3, our word counter, is equal to r0, the program length
	
	cmove r3 .loop r7

end:
	write 'd'
	write 'o'
	write 'n'
	write 'e'
	write 10
	debug 0
	loadLiteral 1024 r7
