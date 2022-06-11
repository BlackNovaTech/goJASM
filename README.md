# gojasm
The jazzy JAS compiler.

[![Release goJASM binaries](https://github.com/BlackNovaTech/goJASM/actions/workflows/release.yml/badge.svg)](https://github.com/BlackNovaTech/goJASM/actions/workflows/release.yml)

## Usage instructions

Invoke the compiler using:
```
$ gojasm input.jas -o output.ijvm
```

To see the selection of useful (and useless) flags use:
```
$ gojasm --help
```

## IJVM extensions

gojasm has a few extensions on the JAS language specification, just for ease of use:

- **numbers**: gojasm allows you to define numbers using either normal (10), 
hex (0x10), octal (012), and binary (0b1010), instead of only the normal form
- **#print macro**: gojasm has a macro, `#print "text to print"` which will be
converted to the corresponding BIPUSH and OUT instructions.

## Custom IJVM configuration

By default we load the equivalent to the Mic-1 default configuration file,
which is defined as:
```
0x10    BIPUSH byte
0x59    DUP
0xA7    GOTO label
0x60    IADD
0x7E    IAND
0x99    IFEQ label
0x9B    IFLT label
0x9F    IF_ICMPEQ label
0x84    IINC var byte
0x15    ILOAD var
0xB6    INVOKEVIRTUAL method
0xB0    IOR
0xAC    IRETURN
0x36    ISTORE var
0x64    ISUB
0x13    LDC_W constant
0x00    NOP
0x57    POP
0x5F    SWAP
0xC4    WIDE
0xFF    HALT
0xFE    ERR
0xFD    OUT
0xFC    IN
```

You can extend this file with as many instructions as you would like.
Every line follows the pattern:
```
opcode name [args...]
```

The following arguments types are available:
 - `byte` (a single byte)
 - `var` (name of a variable)
 - `label` (a label)
 - `constant` (name of a constant)
 - `method` (name of a method)

There can be any numer of arguments.
However, there is a maximum of only one of either `label` or `constant` per operation.

To use your custom configuration, invoke gojasm like follows:
```
$ gojasm -c custom.conf input.jas -o output.jas
```
