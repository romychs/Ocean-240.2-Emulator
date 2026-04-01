; ======================================================
; Ocean-240.2
; Equates for all assembly sources
;
; By Romych 2026-03-10
; ======================================================

	IFNDEF	_EQUATES
	DEFINE	_EQUATES

ASCII_BELL			EQU	0x07                        ; Make Beep
ASCII_BS			EQU	0x08                        ; Move cursor left (Back Space)
ASCII_TAB			EQU	0x09                        ; Move cursor right +8 pos
ASCII_LF			EQU	0x0A                        ; Move cursor down (Line Feed)
ASCII_FF			EQU	0x0C                        ; Move cursor to home (Form Feed)
ASCII_CR			EQU	0x0D                        ; Move gursor to 1st pos (Cariage Return)
ASCII_CAN			EQU	0x18                        ; Move cursor right
ASCII_EM			EQU	0x19                        ; Move cursor up
ASCII_SUB			EQU	0x1A                        ; CTRL-Z - end of text file marker
ASCII_ESC			EQU	0x1B                        ;
ASCII_US			EQU	0x1F                        ; Clear screen
ASCII_SP            EQU 0x20
ASCII_DEL			EQU	0x7F

; ------------------------------------------------------
BELL_PIN			EQU	0x08                        ; DD67 Pin PC3 - "BELL"
; ------------------------------------------------------
CTRL_A              EQU 0x01
CTRL_B              EQU 0x02
CTRL_C              EQU 0x03                        ; Warm boot
CTRL_D              EQU 0x04
CTRL_E              EQU 0x05                        ; Move to beginning of new line (Physical EOL)
CTRL_F              EQU 0x06
CTRL_G              EQU 0x07
CTRL_H              EQU 0x08                        ; Backspace
CTRL_I              EQU 0x09
CTRL_J              EQU 0x0A                        ; LF - Line Feed
CTRL_K              EQU 0x0B
CTRL_L              EQU 0x0C
CTRL_M              EQU 0x0D                        ; CR - Carriage Return
CTRL_N              EQU 0x0E
CTRL_O              EQU 0x0F
CTRL_P              EQU 0x10                        ; turn on/off printer
CTRL_Q              EQU 0x11
CTRL_R              EQU 0x12                        ; Repeat current cmd line
CTRL_S              EQU 0x13                        ; Temporary stop display data to console (aka DC3)
CTRL_T              EQU 0x14
CTRL_U              EQU 0x15                        ; Cancel (erase) current cmd line
CTRL_V              EQU 0x16
CTRL_W              EQU 0x17
CTRL_X              EQU 0x18                        ; Cancel (erase) current cmd line
CTRL_Y              EQU 0x19
CTRL_Z              EQU 0x1A                        ; CTRL-Z - end of text file marker

; ------------------------------------------------------
TRUE				EQU	0xFF
FALSE				EQU	0x00

; ---------------------------------------------------
; FCB Offsets
; ---------------------------------------------------
FCB_LEN				EQU	32                          ; length of FCB
FCB_SHF				EQU	5
FN_LEN              EQU 12                          ; Length of filename in FCB
FCB_DR              EQU 0                           ; Drive. 0 for default, 1-16 for A-P
FCB_FN              EQU 1                           ; Fn - Filename, 7-bit ASCII. The top bits - attributes
FCB_FT              EQU 9                           ; Filetype, 7-bit ASCII.
                                                    ; T1' to T3' have the following
                                                    ; T1' - Read-Only
                                                    ; T2' - System (hidden)
                                                    ; T3' - Archive
FCB_EXT             EQU 12                          ; EX - Set this to 0 when opening a file and then leave it to
                                                    ; CP/M. You can rewind a file by setting EX, RC, S2 and CR to 0.
FCB_S1              EQU 13                          ; S1 - Reserved.
FCB_S2              EQU 14                          ; S2 - Reserved. Bit 7 - wile write flag, [6:0] module number (Extent hi bits)
FCB_RC              EQU 15                          ; RC - Set this to 0 when opening a file and then leave it to CP/M.
FCB_AL              EQU 16                          ; AL - Image of the second half of the directory entry,
                                                    ; containing the file's allocation (which disc blocks it owns).
FCB_CR              EQU 32                          ; CR - Current record within extent. It is usually best to set
                                                    ; this to 0 immediately after a file has been opened and
                                                    ; then ignore it.
FCB_RN              EQU 33                          ; Rn - Random access record number. A 16-bit (with R2 used for overflow)


; ------------------------------------------------------
IRQ_0               EQU 0x01
IRQ_1               EQU 0x02
IRQ_2               EQU 0x04

IRQ_KEYBOARD		EQU	0x02
IRQ_PRINTER			EQU	0x08
IRQ_TIMER			EQU	0x10

KBD_ACK				EQU	0x10
PRINTER_ACK			EQU	0x10

TX_READY			EQU	0x01
RX_READY			EQU	0x02

TMR0_SQWAVE			EQU	0x36

; ------------------------------------------------------

	ENDIF