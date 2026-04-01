# Эмулятор персонального компьютера Океан-240.2

## Пример, кода для работы в среде VSCode.

1. Установите VSCode
2. Установите плагин DeZog для возможности отладки исходного кода в эмуляторе okemu
3. Откройте в VSCode папку examples/hello
4. Запустите эмулятор okemu
5. Запустите отладку через меню Run -> Start debuging в VSCode.
6. Используйте инструменты отладки VSCode для выполнения кода, просмотра регистров процессора и состояния памяти и т.д.

## Рекомендуемые плагины VSCode
- [mborik.z80-macroasm-vscode](https://github.com/mborik/z80-macroasm-vscode) - Раскраска синтаксиса ассемблера Z80, поддержка разных ассемблеров, переименование меток и т.п.
- [maziac.dezog](https://github.com/maziac/DeZog/) - Отладчик кода Z80 с помощью эмуляторов
- [maziac.hex-hover-converter](https://github.com/maziac/hex-hover-converter) - Перевод чисел в другие системы счисления при наведении курсора
- [maziac.z80-instruction-set](https://github.com/maziac/z80-instruction-set) - Подсказки по инструкциям Z80 при наведении курсора (коды, такты, влияние на флаги)

## Пример файла .vscode/tasks.json 
Конфигурация задачи для компиляции исходного кода ассемблером [sjasmplus](https://github.com/z00m128/sjasmplus)
```json
{
    "version": "2.0.0",
    "tasks": [
        {
            "label": "make (sjasmplus)", // так задача будет называться в меню Terminal -> Run task... VSCode
            "type": "shell",
            "command": "sjasmplus",
            "args": [
                "--i8080",	// используем только инструкции КР580ВМ80
                "--sld=main.sld",  // генерируем файл с таблицей символов для отладчика
                "--raw=main.obj",  // сохраняем результат компиляции в двоичный файл
                "--fullpath",
                "main.asm"  // основной файл исходного кода программы
            ],
            "problemMatcher": "$problem-matcher-sjasmplus",
            "group": {
                "kind": "build",
                "isDefault": true
            }
        }
    ]
}
```

## Пример файла .vscode/launch.json
Конфигурация задачи для отладки исходного кода нашей программы с помощью эмулятора okemu. Эмулятор поддерживает подмножество протокола ZRCP (протокол отладки эмулятора [ZEsarUX](https://github.com/chernandezba/zesarux)).
Для отладки нужен плагин DeZog. Подробное описание настроек в [документации плагина](https://github.com/maziac/DeZog/blob/main/documentation/Usage.md#remote-types).

```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "type": "dezog",
            "request": "launch",
            "name": "Simulator", 
            "remoteType": "zrcp", // используем протокол ZRCP для соединения с okemu
            "zrcp": {
                "port": 10001,	  // порт, на котором, по умолчанию, отвечает отладчик okemu 
                                  // если эмулятор запущен на другом компьютере, нужно указать его имя или адрес в параметре "host"
            },

            "sjasmplus": [
                {
                    "path": "main.sld",  // файл с символами для отладки, созданный при компиляции
                },
				// ниже, исходники и файлы для отладки ОС и монитора Океана, если мы хотим видеть их  исходный в VSCode в процессе отладки
                // их можно убрать. Тогда при трассировке вызовов ОС и монитора мы будем видеть дизасм кода.
                {
                    "path": "cpm/cpm.sld",
                    "srcDirs": [
                        "cpm/"
                    ]
                },
                {
                    "path": "mon/monitor.sld",
                    "srcDirs": [
                        "mon/"
                    ]
                }
            ],            
            "smartDisassemblerArgs": {
                "lowerCase": false
            },
            "history": {
                "reverseDebugInstructionCount": 20, // количество инструкций, которые можно шагнуть назад
                "spotCount": 10,
                "codeCoverageEnabled": true //true Если мы хотим видеть, какие участки кода нашего приложения выполнялись
            },
            "startAutomatically": false,
            "commandsAfterLaunch": [
                //"-rmv",                // открыть окно с памятью, на которую указывают 16-разрядные регистры BC,DE,HL,IX,IY
                "-mv 0x100 0x80"         // открыть окно для просмотра памяти с адреса 100h размером 80h байт
            ],
            "rootFolder": "${workspaceFolder}",
            "topOfStack": "stack",
            "loadObjs": [
                {
                    "path": "main.obj", // Этот файл будет загружен в эмулятор
                    "start": "0x0100"   // Файл будет загружен с адреса 100h
                }
            ],
            "execAddress": "0x0100",    // Загруженный код начнет выполняться с адреса 100h
            "smallValuesMaximum": 513,
            "tmpDir": ".tmp"
        }
    ]
}
```
