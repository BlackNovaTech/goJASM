package ijvmasm

import (
	"strings"
	"git.practool.xyz/nova/goJASM/opconf"
	"git.practool.xyz/nova/goJASM/parsers"
)

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
		instruction.wide = true
		method.wide = false
	}
	var bytes uint32 = 1

	for i, token := range params {
		log.Infof("%d -> %s", i, token)
		switch op.Args[i] {
		case opconf.ArgByte:
			val, err := parsers.ParseInt8(token)
			if err != nil {
				asm.Errorf("argument: %s", err.Error())
				return
			}
			instruction.params[i] = int(val)
			bytes += 1
		case opconf.ArgVar:
			idx, ok := method.VarIndex(token)
			if !ok {
				asm.Errorf("argument: Variable not found: `%s`", token)
				return
			}
			instruction.params[i] = idx
			bytes += 1
			if instruction.wide {
				bytes += 1
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
}

func NewInstruction(op *opconf.Operation, N, B uint32) *Instruction {
	return &Instruction{
		op:     op,
		params: make([]int, len(op.Args)),
		N:      N,
		B:      B,
	}
}
