package ijvmasm

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

var (
	// Magic header for the IJVM binaries
	Magic = uint32(0x1DEADFAD)
	// ConstPoolOffset is the byte offset of the constant pool.
	// Stolen from Mic1 emulator
	ConstPoolOffset = uint32(0x10000)
)

// Generate IJVM binary code from the parsed JAS file.
// Returns error if any write fails.
func (asm *Assembler) Generate(out io.Writer) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown generation failure")
			}
		}
	}()

	// Write magic
	mustWrite(out, Magic)
	// Write constant pool offset
	mustWrite(out, ConstPoolOffset)
	// Write constant block size
	mustWrite(out, uint32(len(asm.constants)*4))

	// Write constants
	for _, c := range asm.constants {
		mustWrite(out, c.Value)
	}

	// Write zero (data block memory location)
	mustWrite(out, uint32(0))

	// Write total byte count
	mustWrite(out, uint32(asm.bytes))

	// Generate main
	asm.methods[0].Generate(out)

	for _, m := range asm.methods[1:] {
		mustWrite(out, uint16(m.numparam))
		mustWrite(out, uint16(len(m.vars)-m.numparam))
		m.Generate(out)
	}

	return
}

// GenerateDebugSymbols generates two extra IJVM blocks.
// The first block contains the methods and the second block
// contains the labels for each method in the form 'method#label'
// Each entry in both blocks have the following format:
// <location:u32> <name> '\0'
// Returns error if any write fails
func (asm *Assembler) GenerateDebugSymbols(out io.Writer) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown generation failure")
			}
		}
	}()

	buf := new(bytes.Buffer)
	writer := bufio.NewWriter(buf)

	// Write method headers
	for _, m := range asm.methods {
		mustWrite(writer, m.B)
		mustWrite(writer, []byte(m.name))
		mustWrite(writer, uint8(0))
	}

	writer.Flush()
	outbytes := buf.Bytes()

	mustWrite(out, uint32(0xEEEEEEEE))
	mustWrite(out, uint32(len(outbytes)))
	mustWrite(out, outbytes)

	buf = new(bytes.Buffer)
	writer = bufio.NewWriter(buf)

	// Write labels
	for _, m := range asm.methods {
		for _, l := range m.labels {
			mustWrite(writer, m.B+l.B)
			mustWrite(writer, []byte(fmt.Sprintf("%s#%s", m.name, l.Name)))
			mustWrite(writer, uint8(0))
		}
	}

	writer.Flush()
	outbytes = buf.Bytes()

	mustWrite(out, uint32(0xFFFFFFFF))
	mustWrite(out, uint32(len(outbytes)))
	mustWrite(out, outbytes)
	return
}

func mustWrite(out io.Writer, data interface{}) {
	err := binary.Write(out, binary.BigEndian, data)
	if err != nil {
		panic(err)
	}
}
