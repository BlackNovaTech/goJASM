package ijvmasm

import (
	"io"
	"encoding/binary"
	"errors"
)

var (
	// Magic header, yay
	Magic = uint32(0x1DEADFAD)
	// Byte offset of constant pool stolen from mic1
	ConstPoolOffset = uint32(0x10000)
)

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

func mustWrite(out io.Writer, data interface{}) {
	err := binary.Write(out, binary.BigEndian, data)
	if err != nil {
		panic(err)
	}
}