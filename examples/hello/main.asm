; ==================================================
; Simple CP/M program
; By Romych
; ==================================================
	DEVICE NOSLOT64K
    SLDOPT COMMENT WPMEM, ASSERTION, LOGPOINT

;   INCLUDE "ok240/equates.inc"
    INCLUDE "ok240/bdos.inc"

    OUTPUT  main.com

    ORG  0x100

    LD   B, 10
    LD   SP, stack
AGAIN:
    LD   DE, message
    ; ASSERTION B < 11
    LD   A, 10
    SUB  B
    OR   0x30
    LD   (DE), A
    LD   C, C_WRITESTR
    PUSH BC
    CALL BDOS_ENTER
    POP  BC
    DEC  B  ; LOGPOINT
    JP   NZ, AGAIN
    JP   WARM_BOOT

message:    ; WPMEM, 1, w
    DB "n - Welcome to OK240.2!\r\n$"

    OUTEND

    ;DS 1024
stack    EQU 0xbfc0

    ;DISPLAY "message: EQU\t| ",/H,message


END
