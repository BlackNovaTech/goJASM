package ijvmasm

import (
	"bufio"
	"github.com/op/go-logging"
	"os"
	"strings"
	"fmt"
	"git.practool.xyz/nova/goJASM/opconf"
	"errors"
	"git.practool.xyz/nova/goJASM/parsers"
	"path"
	"regexp"
)

var log = logging.MustGetLogger("ijvmasm")

type Assembler struct {
	opconf *opconf.OpConfig

	fileName string
	scanner  *bufio.Scanner
	line     uint32

	constants []*Constant
	methods   []*Method

	bytes uint32

	parsedMain  bool
	parsedConst bool

	Failed bool
}

func NewAssembler(filepath string, ops *opconf.OpConfig) *Assembler {
	file, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(file)

	return &Assembler{
		opconf:    ops,
		fileName:  path.Base(filepath),
		scanner:   scanner,
		constants: make([]*Constant, 0),
		methods:   make([]*Method, 0),
	}
}

func splitLink(s, sep string) (string, string) {
	x := strings.SplitN(s, sep, 2)
	return x[0], x[1]
}

type Line struct {
	Text string
	N    uint32
}

type Constant struct {
	Name  string
	Value int32
	N     uint32
}

func (asm *Assembler) Parse() (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown assembly failure")
			}
		}
	}()
	for token := asm.next(); token != nil; token = asm.next() {
		log.Debug(asm.Sprintf(token.Text))

		switch token.Text {
		case JASConstantStart:
			asm.constantBlock()
			continue
		case JASMainStart:
			asm.mainBlock()
			continue
		}

		if strings.HasPrefix(token.Text, JASMethodPrefix) {
			if !asm.parsedMain {
				asm.Errorf("Main must be declared before other methods")
				asm.skipUntil(".end-method")
				continue
			}
			asm.methodBlock(strings.TrimPrefix(token.Text, JASMethodPrefix))
		}
	}
	asm.linkMethods()
	return
}

func (asm *Assembler) constantBlock() {
	if asm.parsedConst {
		asm.Errorf("Constant block was already declared")
		asm.skipUntil(JASConstantEnd)
		return
	}

	if asm.parsedMain {
		asm.Errorf("Constant block must appear before methods")
		asm.skipUntil(JASConstantEnd)
		return
	}

	for token := asm.next(); token != nil; token = asm.next() {
		switch token.Text {
		case JASConstantEnd:
			asm.parsedConst = true
			return
		}

		log.Debug(asm.Sprintf(token.Text))

		if token.Text == "" {
			continue
		}

		constant := asm.readConstant(token)
		if constant != nil {
			asm.constants = append(asm.constants, constant)
			log.Infof("Constant registered: %s = %d", constant.Name, constant.Value)
		}

	}
	asm.Panicf("Unexpected end of file\n")
}

func (asm *Assembler) mainBlock() {
	if asm.parsedMain {
		asm.Errorf("Main was already declared. Skipping...")
		asm.skipUntil(JASMainEnd)
		return
	}
	asm.methodBlock("main")
}

func (asm *Assembler) methodBlock(name string) {
	log.Infof("[.%s] Entering method", name)
	method, err := NewMethod(name, asm.line)
	if err != nil {
		panic(err)
	}
	parsedVars := false

	// Instruction parsing
	for token := asm.next(); token != nil; token = asm.next() {
		log.Debug(asm.Sprintf(token.Text))
		switch token.Text {
		case method.end:
			method.LinkLabels()
			asm.methods = append(asm.methods, method)
			log.Infof("Registered method: (%d) %s", len(asm.methods)-1, name)
			return
		case "":
			continue
		case JASVarStart:
			if parsedVars {
				asm.Errorf("Unexpected .var block")
				asm.skipUntil(JASVarEnd)
				continue
			}
			asm.parseVars(method)
			parsedVars = true
			continue

		}

		if !parsedVars {
			parsedVars = true
		}

		instr := token.Text

		if strings.ContainsRune(instr, ':') {
			var labelstr string
			labelstr, instr = splitLink(instr, ":")
			instr = strings.TrimSpace(instr)
			label := &Label{
				strings.TrimSpace(labelstr),
				asm.line,
				method.bytes,
			}
			method.labels = append(method.labels, label)
			log.Infof("[.%s] Registered label: %s@%d", method.name, label.Name, label.B)
		}

		if instr != "" {
			asm.parseInstruction(method, instr)
		}
	}
	asm.Panicf("Unexpected end of file\n")
}

func (asm *Assembler) parseVars(method *Method) {
	for token := asm.next(); token != nil; token = asm.next() {
		log.Debug(asm.Sprintf(token.Text))

		if token.Text == JASVarEnd {
			return
		}

		matched, err := regexp.MatchString("^[a-zA-Z][a-zA-Z0-9_-]*$", token.Text)
		if err != nil {
			asm.Errorf("Variable regex failure: %s", err.Error())
			continue
		}

		if !matched {
			asm.Errorf("Invalid variable name `%s`", token.Text)
			continue
		}

		method.vars = append(method.vars, token.Text)
		log.Infof("[.%s] Registered variable: %s", method.name, token.Text)
	}

	asm.Panicf("Unexpected end of file\n")
	return // Unreachable anyway
}

func (asm *Assembler) readConstant(line *Line) (*Constant) {
	parts := strings.Fields(line.Text)
	if len(parts) < 2 {
		asm.Errorf("constant: Missing constant value")
		return nil
	}

	name, strval := parts[0], parts[1]
	if exists, _, constant := asm.findConstant(name); exists {
		asm.Errorf("constant: Redefinition of constant `%s` from line %d", name, constant.N)
		return nil
	}

	word, err := parsers.ParseInt32(strval)
	if err != nil {
		asm.Errorf("constant: %s", err.Error())
		return nil
	}

	return &Constant{
		N:     line.N,
		Name:  name,
		Value: word,
	}

}

func (asm *Assembler) findConstant(name string) (bool, int, *Constant) {
	for index, constant := range asm.constants {
		if constant.Name == name {
			return true, index, constant
		}
	}
	return false, -1, nil
}

func (asm *Assembler) next() *Line {
	if asm.scanner.Scan() {
		asm.line += 1
		return &Line{
			N:    asm.line,
			Text: strings.TrimSpace(strings.SplitN(asm.scanner.Text(), "//", 2)[0]),
		}
	}
	return nil
}

func (asm *Assembler) Sprintf(format string, args ...interface{}) string {
	vars := append([]interface{}{asm.fileName, asm.line}, args...)

	return fmt.Sprintf("%s:%d > "+format, vars...)
}

func (asm *Assembler) Errorf(format string, args ...interface{}) {
	asm.Failed = true
	log.Errorf(asm.Sprintf(format, args...))
}

func (asm *Assembler) Panicf(format string, args ...interface{}) {
	asm.Failed = true
	panic(errors.New(asm.Sprintf(format, args...)))
}

func (asm *Assembler) skipUntil(pattern string) {
	for token := asm.next(); token != nil; token = asm.next() {
		if token.Text == pattern {
			return
		}
		log.Warning(asm.Sprintf("|skip| %s", token.Text))
	}
}

func (asm *Assembler) linkMethods() (ok bool) {
	ok = true
	if len(asm.methods) == 0 {
		asm.Errorf("linker: No main found")
		return false
	}
	asm.bytes = asm.methods[0].bytes
	for i, method := range asm.methods[1:] {

		if exists, _, constant := asm.findConstant(method.name); exists {
			asm.Errorf("linker: Method constant name conflict. `%s` already defined on line %d", method.name, constant.N)
			return
		}

		mconst := &Constant{
			N:     method.N,
			Name:  method.name,
			Value: int32(asm.bytes),
		}

		asm.constants = append(asm.constants, mconst)
		log.Infof("Method #%d line %d placed at %d", i, method.N, asm.bytes)
		asm.bytes += method.bytes + 4
	}

	for _, method := range asm.methods {
		ok = ok && method.LinkMethods(asm)
	}
	return
}
