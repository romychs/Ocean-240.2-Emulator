; =======================================================
; Ocean-240.2
; Equates to call Monitor v8 functions
;
; Created by Romych 2026-02-16
; =======================================================

    IFNDEF      _MON_INC
    DEFINE      _MON_INC

mon_start           EQU 0xE000
mon_hexb            EQU 0xE003
non_con_status      EQU 0xE006
mon_con_in          EQU 0xE009
mon_con_out         EQU 0xE00C
mon_serial_in       EQU 0xE00F
mon_serial_out      EQU 0xE012
mon_char_print      EQU 0xE015
mon_tape_read       EQU 0xE018
mon_tape_write      EQU 0xE01B
mon_ram_disk_read   EQU 0xE01E
mon_ram_disk_write  EQU 0xE021
mon_tape_read_ram   EQU 0xE024
mon_tape_write_ram  EQU 0xE027
mon_tape_wait       EQU 0xE02A
mon_tape_detect     EQU 0xE02D
mon_read_floppy     EQU 0xE030
mon_write_floppy    EQU 0xE033
mon_out_str_z       EQU 0xE036

    ENDIF