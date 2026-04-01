# Эмулятор персонального компьютера Океан-240.2

![Icon](img/IconBig.png?raw=true "AppIcon")

## Зачем?

Эмулятор был разработан для удобства реверс-инжененринга программ монитора и приложений для этого старого компьютера.
Поэтому и имеет такой минималистичный интерфейс.

Если нужен более точный эмулятор Океана-240.2 и других компьютеров на безе процессора КР580ВМ80, рекомендую проект [Emu80](https://emu80.org/).

## Особенности реализации

- Микросхемы КР580ВВ55, ВИ53, ВИ51, ВТ59, КР1818ВГ93 эмулируются в степени, достаточной для работы операционной системы и монитора.
- В отличие от оригинала, который использует CPU КР580ВМ80, в эмуляторе использована эмуляция Z80.
  Эмулятор Z80 у меня уже был написан, поэтому использовал его, возможно, позже, обрежу его до i8080.
  Из за меньшего количества тактов у некоторых инструкций Z80, эмуляция работает несколько быстрее оригинала.
- Эмулируются 2 дисковода с дисками 720К, такой вариант используется с обазами ROM CP/M и Монитора R8.
- Работает под современными ОС Windows и Linux. Используется фреймворк [Fyne](https://fyne.io/), что позволяет скомпилировать код и под другие платформы, но я не пробовал.
- Поддерживает ZRCP - протокол отладки эмулятора [ZEsarUX](https://github.com/chernandezba/zesarux). Это позволяет использовать среду разработки VSCode с плагином DeZog
для отладки исходного кода прямо в эмуляторе.

## Возможности отладки

Работают все функции плагина Dezog и протокола: [документация плагина](https://github.com/maziac/DeZog/blob/main/documentation/Usage.md#remote-types).

- Выполнение кода по шагам, в том числе и назад (с ограничениями плагина DeZog)
- Просмотр и изменение памяти
- Просмотр и изменение значения регистров процессора
- Условные и безусловные брейкпоинты
- ASSERTION - остановка по несоблюдению указанного условия
- WPMEM - остановка при обращении к указанным ячейкам памяти (можно указать тип обращения r|w|rw)
- CodeCoverage - в коде, цветом помечаются строки, выполненные процессором.

![Debug](img/Debug.jpg?raw=true "Debug in VSCode")

Можно отлаживать код и без VSCode, подключившись телнетом к порту отладчика, по умолчанию это localhost:10001, далее, можно передавать отладчику команды.
Их список доступен по команде help.

    telnet localhost 10001

### Список доступных команд отладчика

    about                   Shows about message
    clear-membreakpoints    Clear all memory breakpoints
    close-all-menus         Close all visible dialogs
    cpu-code-coverage       Sets cpu code coverage parameters
    cpu-history             Runs cpu history actions
    cpu-step                Run single opcode cpu step
    disable-breakpoint      Disable specific breakpoint
    disable-breakpoints     Disable all breakpoints
    disassemble             Disassemble at address
    enable-breakpoint       Enable specific breakpoint
    enable-breakpoints      Enable breakpoints
    enter-cpu-step          Enter cpu step to step mode
    evaluate                Evaluate expression
    exit-cpu-step           Exit cpu step to step mode
    extended-stack          Sets extended stack parameters, which allows you to see what kind of values are in the stack
    get-cpu-frequency       Get cpu frequency in HZ
    get-current-machine     Returns current machine name
    get-machines            Returns list of emulated machines
    get-membreakpoints      Get memory breakpoints list
    get-memory-pages        Returns current state of memory pages
    get-os                  Shows emulator operating system
    get-registers           Get CPU registers
    get-tstates-partial     Get the t-states partial counter
    get-tstates             Get the t-states counter
    get-version             Shows emulator version
    hard-reset-cpu          Hard resets the machine
    help                    Shows help screen or command help
    hexdump                 Dumps memory at address, showing hex and ascii
    load-binary             Load binary file "file" at address "addr" with length "len", on the current memory zone
    quit                    Closes connection
    read-memory             Dumps memory at address
    reset-tstates-partial   Resets the t-states partial counter
    run                     Run cpu when on cpu step mode
    save-binary             Save binary file "file" from address "addr" with length "len", from the current memory zone
    set-breakpoint          Sets a breakpoint at desired index entry with condition
    set-breakpointaction    Sets a breakpoint action at desired index entry
    set-breakpointpasscount Set pass count for breakpoint
    set-debug-settings      Set debug settings on remote command protocol
    set-machine             Set machine
    set-membreakpoint       Sets a memory breakpoint starting at desired address entry for type
    set-register            Changes register value
    snapshot-load           Loads a snapshot
    snapshot-save           Saves a snapshot
    write-memory            Writes a sequence of bytes starting at desired address on memory
    write-port              Writes value at port
