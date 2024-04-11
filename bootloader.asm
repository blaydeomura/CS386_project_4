; Store first two bytes of input into r0
main:
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

	loadLiteral 1024 r7