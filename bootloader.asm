; Store first two bytes of input into r0
main:
	read r0							; read 1/2 bytes
	shl r0 8 r0						; shift left 8bits
	read r0							; read 2/2 bytes

	loadLiteral 1024 r5				; r2: 1024, just preserves the amount we need to buffer the program by

	loadLiteral 0 r2				; make r2 to 0

; Next: Keep reading from input into r1 until we get 0
loop:
	; First: Read a word (4 bytes) into r1, r2, r3, and r4
	;        - each read is 8 bytes
	loadLiteral 0 r3				; make r3 to 0
	
	read r1							; read r1
	shl r1 24 r1					; shift left r1 to 24 and store that into r1
	or r3 r1 r3						; r3 || r1 and store that into r3
	
	read r1							; read r1
	shl r1 16 r1					; shift r1 left 16 
	or r3 r1 r3						; combine r1 into r3

	read r1							; repeat
	shl r1 8 r1
	or r3 r1 r3
	
	read r1							; no need to shift this time
	or r3 r1 r3

	store r3 r5						; store the word into memory address

	load r5 r3						
	
	add r5 1 r5
	add r2 1 r2

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
