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
- Эмулируются 2 дисковода с дисками 720К, такой вариант используется с CP/M и Монитором R8.
- Работает под современными ОС Windows и Linux. Используется фреймворк [Fyne](https://fyne.io/), что позволяет скомпилировать код и под другие платформы, но я не пробовал.
- Поддерживает ZRCP - протокол отладки эмулятора [ZEsarUX](https://github.com/chernandezba/zesarux)). Это позволяет использовать среду разработки VSCode с плагином DeZog
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

    __about__                   Shows about message
    __clear-membreakpoints__    Clear all memory breakpoints
    __close-all-menus__         Close all visible dialogs
    __cpu-code-coverage__       Sets cpu code coverage parameters
    __cpu-history__             Runs cpu history actions
    __cpu-step__                Run single opcode cpu step
    __disable-breakpoint__      Disable specific breakpoint
    __disable-breakpoints__     Disable all breakpoints
    __disassemble__             Disassemble at address
    __enable-breakpoint__       Enable specific breakpoint
    __enable-breakpoints__      Enable breakpoints
    __enter-cpu-step__          Enter cpu step to step mode
    __evaluate__                Evaluate expression
    __exit-cpu-step__           Exit cpu step to step mode
    __extended-stack__          Sets extended stack parameters, which allows you to see what kind of values are in the stack
    __get-cpu-frequency__       Get cpu frequency in HZ
    __get-current-machine__     Returns current machine name
    __get-machines__            Returns list of emulated machines
    __get-membreakpoints__      Get memory breakpoints list
    __get-memory-pages__        Returns current state of memory pages
    __get-os__                  Shows emulator operating system
    __get-registers__           Get CPU registers
    __get-tstates-partial__     Get the t-states partial counter
    __get-tstates__             Get the t-states counter
    __get-version__             Shows emulator version
    __hard-reset-cpu__          Hard resets the machine
    __help__                    Shows help screen or command help
    __hexdump__                 Dumps memory at address, showing hex and ascii
    __load-binary__             Load binary file "file" at address "addr" with length "len", on the current memory zone
    __quit__                    Closes connection
    __read-memory__             Dumps memory at address
    __reset-tstates-partial__   Resets the t-states partial counter
    __run__                     Run cpu when on cpu step mode
    __save-binary__             Save binary file "file" from address "addr" with length "len", from the current memory zone
    __set-breakpoint__          Sets a breakpoint at desired index entry with condition
    __set-breakpointaction__    Sets a breakpoint action at desired index entry
    __set-breakpointpasscount__ Set pass count for breakpoint
    __set-debug-settings__      Set debug settings on remote command protocol
    __set-machine__             Set machine
    __set-membreakpoint__       Sets a memory breakpoint starting at desired address entry for type
    __set-register__            Changes register value
    __snapshot-load__           Loads a snapshot
    __snapshot-save__           Saves a snapshot
    __write-memory__            Writes a sequence of bytes starting at desired address on memory
    __write-port__              Writes value at port
