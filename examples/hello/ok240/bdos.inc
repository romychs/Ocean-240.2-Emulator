; =======================================================
; Ocean-240.2
; CP/M v2.2. Equates to call BDOS functions
;
; Created by Romych 2026-02-16
; =======================================================


    IFNDEF   _BDOS_INC
    DEFINE   _BDOS_INC

; ------------------------------------------------------
; CP/M Entries and variables
; ------------------------------------------------------
WARM_BOOT           EQU 0x0000                      ; Jump warm_boot (Restart)
IO_BYTE             EQU 0x0003                      ; Input/Output mapping
CUR_USER_DRV        EQU 0x0004                      ; [7:4] - curent user, [3:0] - current drive
BDOS_ENTER          EQU 0x0005                      ; Jump bdos (CALL 5 to make CP/M requests)
FCB1                EQU 0x005c                      ; Default FCB, 16 bytes
FCB2                EQU 0x006c

DMA_BUFFER          EQU 0x0080                      ; Default "DMA" 128 bytes buffer
CMD_LINE_LEN        EQU 0x0080                      ; command line character count
CMD_LINE            EQU 0x0081                      ; command line buffer
TPA                 EQU 0x0100                      ; start of program


; ------------------------------------------------------
; CP/M BDOS Functions
; ------------------------------------------------------
P_TERMCPM           EQU 0                           ; System Reset
C_READ              EQU 1                           ; Console input
C_WRITE             EQU 2                           ; Console output
A_READ              EQU 3                           ; Auxiliary (Reader) input
A_WRITE             EQU 4                           ; Auxiliary (Punch) output
L_WRITE             EQU 5                           ; Printer output
C_RAWIO             EQU 6                           ; Direct console I/O
A_STATIN            EQU 7                           ; Auxiliary Input status
A_STATOUT           EQU 8                           ; Auxiliary Output status
C_WRITESTR          EQU 9                           ; Output string
C_READSTR           EQU 10                          ; Buffered console input
C_STAT              EQU 11                          ; Console status
S_BDOSVER           EQU 12                          ; Return version number
DRV_ALLRESET        EQU 13                          ; Reset discs
DRV_SET             EQU 14                          ; Select disc
F_OPEN              EQU 15                          ; Open file
F_CLOSE             EQU 16                          ; Close file
F_SFIRST            EQU 17                          ; search for first
F_SNEXT             EQU 18                          ; search for next
F_DELETE            EQU 19                          ; delete file
F_READ              EQU 20                          ; read next record
F_WRITE             EQU 21                          ; write next record
F_MAKE              EQU 22                          ; create file
F_RENAME            EQU 23                          ; Rename file
DRV_LOGINVEC        EQU 24                          ; Return bitmap of logged-in drives
DRV_GET             EQU 25                          ; Return current drive
F_DMAOFF            EQU 26                          ; Set DMA address
DRV_ALLOCVEC        EQU 27                          ; Return address of allocation map
DRV_SETRO           EQU 28                          ; Software write-protect current disc
DRV_ROVEC           EQU 29                          ; Return bitmap of read-only drives
F_ATTRIB            EQU 30                          ; set file attributes
DRV_DPB             EQU 31                          ; get DPB address
F_USERNUM           EQU 32                          ; get/set user number
F_READRAND          EQU 33                          ; Random access read record
F_WRITERAND         EQU 34                          ; Random access write record
F_SIZE              EQU 35                          ; Compute file size
F_RANDREC           EQU 36                          ; Update random access pointer
DRV_RESET           EQU 37                          ; Selectively reset disc drives
DRV_ACCESS          EQU 38                          ; Access drives
DRV_FREE            EQU 39                          ; Free drive
F_WRITEZF           EQU 40                          ; Write random with zero fill

    ENDIF
