package ijvmasm

import (
	"strings"

	"github.com/BlackNovaTech/goJASM/opconf"
	"github.com/BlackNovaTech/goJASM/parsers"
)

// Instruction is a single instruction in the instruction stream.
type Instruction struct {
	op     *opconf.Operation
	params []int
	label  string
	N      uint32
	B      uint32

	wide       bool
	linkLabel  bool
	linkMethod bool
}

// Parses a single instruction string for the given method
func (asm *Assembler) parseInstruction(method *Method, instr string) {
	tokens := strings.Fields(instr)
	opname := tokens[0]
	params := tokens[1:]
	op := asm.opconf.GetOp(opname)
	if op == nil {
		asm.Errorf("Undefined instruction `%s`", instr)
		return
	}

	if len(params) != len(op.Args) {
		asm.Errorf("Mismatched argument count, expected %d, got %d", len(op.Args), len(params))
		return
	}

	instruction := NewInstruction(op, asm.line, method.bytes)
	if method.wide {
		log.Infof("[.%s] Operation widened", method.name, )
		instruction.wide = true
		method.wide = false
	}
	var bytes uint32 = 1

	for i, token := range params {
		log.Debugf("arg %d -> %s", i, token)
		switch op.Args[i] {
		case opconf.ArgByte:
			var val int8
			var err error
			if strings.HasPrefix(token, "'") {
				val, err = parsers.ParseChar(token)
			} else {
				val, err = parsers.ParseInt8(token)
			}
			if err != nil {
				asm.Errorf("argument: %s", err.Error())
				return
			}
			instruction.params[i] = int(val)
			bytes++
		case opconf.ArgVar:
			idx, ok := method.VarIndex(token)
			if !ok {
				asm.Errorf("argument: Variable not found: `%s`", token)
				return
			}
			instruction.params[i] = idx
			bytes++
			if instruction.wide {
				bytes++
			} else if idx > 0xFF {
				if asm.AutoWide {
					method.AppendInst(NewInstruction(asm.opconf.GetOp(OperationWide), asm.line, method.bytes))
					bytes++
					instruction.B++
					instruction.wide = true
					log.Infof("[.%s] Auto widened instruction", method.name)
				} else {
					asm.Errorf("argument: Variable index out of range: `%s` (%d > %d)", token, idx, 0xFF)
					return
				}
			}
		case opconf.ArgLabel:
			instruction.label = token
			instruction.linkLabel = true
			bytes += 2
		case opconf.ArgConst:
			ok, idx, _ := asm.findConstant(token)
			if !ok {
				asm.Errorf("argument: Constant not found: `%s`", token)
				return
			}
			instruction.params[i] = idx
			bytes += 2
		case opconf.ArgMethod:
			instruction.label = token
			instruction.linkMethod = true
			bytes += 2
		default:
			asm.Errorf("Not implemented")
		}
	}

	method.AppendInst(instruction)
	method.bytes += bytes
	log.Infof("[.%s] Registered instruction: %s (%d)", method.name, instruction.op.Name, len(instruction.params))

	if instruction.op.Name == OperationWide {
		method.wide = true
	}
}

// NewInstruction creates a new Instruction based on the given Operation, line number, and byte number.
func NewInstruction(op *opconf.Operation, N, B uint32) *Instruction {
	return &Instruction{
		op:     op,
		params: make([]int, len(op.Args)),
		N:      N,
		B:      B,
	}
}
