package main

import (
	"bufio"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strings"
)

func main() {
	scanner := scannerFromFile("test2.jas")
	asm := Assembler{
		scanner:   scanner,
		constants: make([]*Constant, 0),
		methods:   make([]*Method, 0),
	}
	defer asm.recoverParseError()
	asm.lineTokens()
}

func scannerFromFile(path string) *bufio.Scanner {
	f, err := os.Open(path)
	if err != nil {
		fmt.Println("Could not open file")
	}
	return bufio.NewScanner(f)
}

type Assembler struct {
	scanner     *bufio.Scanner
	line        uint32
	constants   []*Constant
	methods     []*Method
	parsedMain  bool
	parsedConst bool
}

type Instruction struct {
	name string
	args []string
	N    uint32
}

type Method struct {
	name         string
	args         []string
	vars         []string
	params		 []string
	instructions []*Instruction
	labels       []*Label
	end          string
}
func splitLink(s, sep string) (string, string) {
	x := strings.SplitN(s, sep, 2)
	return x[0], x[1]
}

func NewMethod(nameParam string) (*Method, error) {
	end := ".end-method"
	params := []string{}
	name := "main"
	if nameParam == "main" {
		end = ".end-main"
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
		params = make([]string, len(paramlst))
		for i,p := range paramlst {
			params[i] = strings.TrimSpace(p)
		}
	}

	return &Method{
		name:         name,
		args:         make([]string, 0),
		vars:         make([]string, 0),
		instructions: make([]*Instruction, 0),
		labels:       make([]*Label, 0),
		end:          end,
		params:       params,
	}, nil

}

type Line struct {
	Text string
	N    uint32
}

type Label struct {
	*Line
}

type Constant struct {
	Name  string
	Value int32
	Line  uint32
}

func (asm *Assembler) lineTokens() {
	for token := asm.next(); token != nil; token = asm.next() {
		switch token.Text {
		case ".constant":
			asm.constantBlock()
			continue
		case ".main":
			asm.mainBlock()
			continue
		}
		if strings.HasPrefix(token.Text, ".method ") {
			asm.methodBlock(strings.TrimPrefix(token.Text, ".method "))
		} else {
			asm.Logf(token.Text)
		}
	}
}

func (asm *Assembler) constantBlock() {
	asm.Logf("[CONST] .constant")

	if asm.parsedConst {
		asm.Panicf("[CONST] Constant block was already declared")
	}

	if asm.parsedMain {
		asm.Panicf("[CONST] Constant block must appear before methods")
	}

	for token := asm.next(); token != nil; token = asm.next() {
		switch token.Text {
		case ".end-constant":
			asm.parsedConst = true
			asm.Logf("[CONST] .end-constant")
			return
		case "":
			asm.Logf("[CONST] ")
			continue
		}

		constant, err := asm.readConstant(token)
		if err != nil {
			panic(err)
		}

		asm.constants = append(asm.constants, constant)
		asm.Logf("[CONST] New constant %s=%d", constant.Name, constant.Value)

	}
	asm.Panicf("Unexpected end of file\n")
}

func (asm *Assembler) mainBlock() {
	//asm.Logf("[MAIN] .main")

	if asm.parsedMain {
		panic(errors.New(asm.Sprintf("Main was already declared")))
	}

	asm.methodBlock("main")

}

func (asm *Assembler) methodBlock(name string) *Method {
	method, err := NewMethod(name)
	if err != nil {
		panic(err)
	}
	asm.Logf("[.%s] %s", method.name, name)


	for token := asm.next(); token != nil; token = asm.next() {
		switch token.Text {
		case method.end:
			asm.Logf("[.%s] %s", method.name, token.Text)
			return method
		case "":
			asm.Logf("[.%s] %s", method.name, token.Text)
			continue
		}

		instr := token.Text

		if strings.ContainsRune(instr, ':') {
			var label string
			label, instr = splitLink(instr,":")
			label = strings.TrimSpace(label)
			asm.Logf("[.%s] Label: %s", method.name, label)
		}
		asm.Logf("[.%s] Instr: %s", method.name, instr)

	}
	asm.Panicf("Unexpected end of file\n")
	return nil // Unreachable anyway
}

func (asm *Assembler) readConstant(line *Line) (*Constant, error) {
	parts := strings.Fields(line.Text)
	if len(parts) != 2 {
		return nil, asm.Errorf("Expected constant")
	}
	name, strval := parts[0], parts[1]
	if exists, _, constant := asm.findConstant(name); exists {
		return nil, asm.Errorf("Redefinition of constant `%s` from line %d", name, constant.Line)
	}

	word, err := asm.parseWord(strval)
	if err != nil {
		return nil, err
	}

	return &Constant{
		Line:  line.N,
		Name:  name,
		Value: word,
	}, nil

}

func (asm *Assembler) parseWord(strval string) (int32, error) {
	bigval, ok := big.NewInt(-1).SetString(strval, 0)
	if !ok {
		return -1, asm.Errorf("Invalid constant literal `%s`", strval)
	}
	if bigval.BitLen() >= 32 {
		return -1, asm.Errorf("Value out of range: `%s`", strval)
	}

	return int32(bigval.Int64()), nil

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
			Text: strings.TrimSpace(strings.SplitN(asm.scanner.Text(), "//",2)[0]),
		}
	}
	return nil
}

func (asm *Assembler) Sprintf(format string, args ...interface{}) string {
	vars := append([]interface{}{asm.line}, args...)

	return fmt.Sprintf("[%3d] "+format, vars...)
}

func (asm *Assembler) Logf(format string, args ...interface{}) {
	fmt.Println(asm.Sprintf(format, args...))
}

func (asm *Assembler) Errorf(format string, args ...interface{}) error {
	return errors.New(asm.Sprintf(format, args...))
}

func (asm *Assembler) Panicf(format string, args ...interface{}) {
	panic(asm.Errorf(format, args...))
}

func (asm *Assembler) recoverParseError() {
	if r := recover(); r != nil {
		if err, ok := r.(error); ok {
			fmt.Println("\n\n[ERROR] Assembly failed")
			fmt.Println(err.Error())
			os.Exit(1)
		} else {
			panic(r)
		}
	}
}
