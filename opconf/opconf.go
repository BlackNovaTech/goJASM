package opconf

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/BlackNovaTech/gojasm/parsers"
	"github.com/sirupsen/logrus"
)

// ArgType represents the type of an argument
type ArgType int8

const (
	// ArgByte represents a byte argument
	ArgByte ArgType = iota
	// ArgLabel represents a label argument, that requires further linking later on
	ArgLabel
	// ArgVar represents a variable argument that is converted into a variable index
	ArgVar
	// ArgMethod represents a method argument, that requires further linking later on.
	ArgMethod
	// ArgConst represents a constant argument, that is converted into a constant index
	ArgConst
)

// OpConfig represents the full available suite of operations the compiler understands
type OpConfig struct {
	fileName   string
	scanner    *bufio.Scanner
	line       uint32
	operations map[string]*Operation
	failed     bool
}

// Operation is a single o peration that the compiler can parse
type Operation struct {
	// Name is a nice representable name (e.g. BIPUSH)
	Name string
	// Opcode is the byte representation (e.g. 0x10)
	Opcode uint8
	// Args is a variable list of ArgType that represents the arguments the operation takes
	Args []ArgType
}

// NewOpConfigFromPath generates an OpConf from the given file
func NewOpConfigFromPath(filepath string) *OpConfig {
	file, err := os.Open(filepath)
	if err != nil {
		logrus.Fatal(err)
	}

	config := NewOpConfig(file, path.Base(filepath))

	return config
}

// NewDefaultOpConfig generates an OpConf from the default set of instructions
func NewDefaultOpConfig() *OpConfig {
	config := NewOpConfig(strings.NewReader(defaultConfig), "default")

	return config
}

// NewOpConfig generates an OpConf from the given source, optionally with the given name
func NewOpConfig(read io.Reader, name string) *OpConfig {
	scanner := bufio.NewScanner(read)

	config := &OpConfig{
		scanner:    scanner,
		fileName:   name,
		operations: make(map[string]*Operation),
	}

	config.parse()
	if config.failed {
		logrus.Fatal("OpConf parse failed")
	}

	return config
}

// GetOp retrieves the operation corresponding to the given name
func (cfg *OpConfig) GetOp(opname string) *Operation {
	if op, ok := cfg.operations[opname]; ok {
		return op
	}
	return nil
}

// Parses a configuration file
func (cfg *OpConfig) parse() {
	ophash := make(map[uint8]bool)
	for tokens := cfg.next(); tokens != nil; tokens = cfg.next() {
		logrus.Debug(cfg.Sprintf(strings.Join(tokens, " ")))

		if len(tokens) == 0 {
			continue
		}

		if len(tokens) < 2 {
			cfg.Errorf("Missing operation name")
			continue
		}

		opcode, err := parsers.ParseUint8(tokens[0])
		if err != nil {
			cfg.Errorf("opcode: %s", err.Error())
			continue
		}

		if _, ok := ophash[opcode]; ok {
			cfg.Errorf("Duplicate opcode `%2X`", opcode)
			continue
		}

		opname := strings.ToUpper(tokens[1])

		if _, ok := cfg.operations[opname]; ok {
			cfg.Errorf("Duplicate operation `%s`", opname)
			continue
		}

		args := make([]ArgType, len(tokens[2:]))

		// it got arguments!
		for i, s := range tokens[2:] {
			args[i], err = parseArg(s)
			if err != nil {
				cfg.Errorf("argument: %s", err.Error())
			}
		}
		op := &Operation{
			Name:   opname,
			Opcode: opcode,
			Args:   args,
		}

		ophash[opcode] = true
		cfg.operations[opname] = op
		logrus.Debugf("Operation registered: %2X -> %s (%d)", op.Opcode, op.Name, len(op.Args))
	}
}

// Next token in config file
func (cfg *OpConfig) next() []string {
	if cfg.scanner.Scan() {
		cfg.line++
		return strings.Fields(strings.SplitN(cfg.scanner.Text(), "//", 2)[0])
	}
	return nil
}

// Parses argument type
func parseArg(str string) (arg ArgType, err error) {
	switch strings.ToLower(str) {
	case "byte":
		arg = ArgByte
	case "label":
		arg = ArgLabel
	case "var":
		arg = ArgVar
	case "method":
		arg = ArgMethod
	case "constant":
		arg = ArgConst
	default:
		err = fmt.Errorf("Unknown argument type `%s`", str)
	}
	return
}

// Sprintf formats given arguments, but prepends filename and line number.
func (cfg *OpConfig) Sprintf(format string, args ...interface{}) string {
	vars := append([]interface{}{cfg.fileName, cfg.line}, args...)
	return fmt.Sprintf("%s:%d > "+format, vars...)
}

// Errorf sets the failed flag of the assembler, and then logs an error
// using OpConfig#Sprintf
func (cfg *OpConfig) Errorf(format string, args ...interface{}) {
	cfg.failed = true
	logrus.Error(cfg.Sprintf(format, args...))

}
