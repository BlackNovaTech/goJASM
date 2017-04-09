package ijvmasm

import (
	"strings"
	"errors"
	"git.practool.xyz/nova/goJASM/opconf"
	"io"
)

type Method struct {
	name         string
	vars         []string
	instructions []*Instruction
	labels       []*Label
	end          string

	bytes        uint32
	numparam     int

	N uint32

	wide bool
}

type Label struct {
	Name string
	N    uint32
	B    uint32
}

func NewMethod(nameParam string, N uint32) (*Method, error) {
	end := JASMethodEnd
	params := []string{}
	name := "main"
	if nameParam == "main" {
		end = JASMainEnd
	} else {
		if !strings.ContainsRune(nameParam, '(') {
			return nil, errors.New("Invalid method declaration. Missing opening parenthesis")
		}
		rawName, paramstr := splitLink(nameParam, "(")
		name = strings.TrimSpace(rawName)
		if !strings.ContainsRune(paramstr, ')') {
			return nil, errors.New("Invalid method declaration. Missing closing parenthesis")
		}
		paramstr, junk := splitLink(paramstr, ")")
		if strings.TrimSpace(junk) != "" {
			return nil, errors.New("Invalid method declaration. Characters remaining after parameter list")
		}


		paramlst := strings.Split(paramstr, ",")
		// BS empty parameter list
		if paramstr == "" {
			paramlst = []string{}
		}

		params = make([]string, len(paramlst)+1)
		params[0] = "LINK PTR"
		for i, p := range paramlst {
			params[i+1] = strings.TrimSpace(p)
		}
	}

	return &Method{
		name:         name,
		vars:         params,
		numparam: len(params),
		instructions: make([]*Instruction, 0),
		labels:       make([]*Label, 0),
		end:          end,
		N:            N,
	}, nil

}

func (m *Method) AppendInst(inst *Instruction) {
	m.instructions = append(m.instructions, inst)
}

func (m *Method) VarIndex(str string) (int, bool) {
	for i, p := range m.vars {
		if p == str {
			return i, true
		}
	}
	return -1, false
}

func (m *Method) findLabel(name string) (bool, int, *Label) {
	for i, label := range m.labels {
		if label.Name == name {
			return true, i, label
		}
	}
	return false, -1, nil
}

func (m *Method) LinkLabels() (ok bool) {
	ok = true
	for _, inst := range m.instructions {
		if !inst.linkLabel {
			continue
		}
		for j, argType := range inst.op.Args {
			if argType == opconf.ArgLabel {
				found, _, lbl := m.findLabel(inst.label)
				if !found {
					log.Errorf("[.%s] Undefined label `%s` at line %d", m.name, inst.label, inst.N)
					ok = false
				}
				inst.params[j] = int(lbl.B) - int(inst.B)
				log.Debugf("[.%s] Linking label, line %d: @%d -> %s@%d, offset = %d",
					m.name, inst.N, inst.B, lbl.Name, lbl.B, inst.params[j])
			}
		}
	}
	return
}

func (m *Method) LinkMethods(asm *Assembler) (ok bool) {
	ok = true
	for _, inst := range m.instructions {
		if !inst.linkMethod {
			continue
		}
		for j, argType := range inst.op.Args {
			if argType == opconf.ArgMethod {
				found, idx, mtd := asm.findConstant(inst.label)
				if !found {
					log.Errorf("[.%s] Undefined method `%s` at line %d", m.name, inst.label, inst.N)
					ok = false
				}
				inst.params[j] = idx
				log.Debugf("[.%s] Linking method, line %d: %s -> %d",
					m.name, inst.N, mtd.Name, inst.params[j])			}
		}
	}
	return
}

func (m *Method) Generate(out io.Writer) {
	for _, inst := range m.instructions {
		mustWrite(out, inst.op.Opcode)
		for i, arg := range inst.op.Args {
			switch arg {
			case opconf.ArgByte:
				mustWrite(out, uint8(inst.params[i]))
			case opconf.ArgVar:
				if inst.wide {
					mustWrite(out, uint16(inst.params[i]))
				} else {
					mustWrite(out, uint8(inst.params[i]))
				}
			case opconf.ArgLabel:
				fallthrough
			case opconf.ArgConst:
				fallthrough
			case opconf.ArgMethod:
				mustWrite(out, uint16(inst.params[i]))

			default:
				panic("Unimplemented")
			}
		}
	}
}
