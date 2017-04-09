package opconf

import (
	"bufio"
	"strings"
	"fmt"
	"github.com/op/go-logging"
	"os"
	"git.practool.xyz/nova/goJASM/parsers"
	"errors"
	"path"
	"io"
)

var log = logging.MustGetLogger("opconf")

type ArgType int8

const (
	ArgByte ArgType = iota
	ArgLabel
	ArgVar
	ArgMethod
	ArgConst
)

type OpConfig struct {
	fileName    string
	scanner     *bufio.Scanner
	line        uint32
	operations  map[string]*Operation
	failed      bool
}

type Operation struct {
	Name string
	Opcode uint8
	Args []ArgType
}

func NewOpConfigFromPath(filepath string) *OpConfig {
	file, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}

	config := NewOpConfig(file, path.Base(filepath))

	return config
}

func NewDefaultOpConfig() *OpConfig {
	config := NewOpConfig(strings.NewReader(DefaultConfig), "default")

	return config
}

func NewOpConfig(read io.Reader, name string) *OpConfig {
	scanner := bufio.NewScanner(read)

	config := &OpConfig{
		scanner:    scanner,
		fileName:   name,
		operations: make(map[string]*Operation),
	}

	config.parse()
	if config.failed {
		log.Fatal("OpConf parse failed")
	}

	return config
}

func (cfg *OpConfig) GetOp(opname string) *Operation {
	if op, ok := cfg.operations[opname]; ok {
		return op
	}
	return nil
}

func (cfg *OpConfig) parse() {
	ophash := make(map[uint8]bool)
	for tokens := cfg.next(); tokens != nil; tokens = cfg.next() {
		log.Debug(cfg.Sprintf(strings.Join(tokens, " ")))

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

		opname := tokens[1]

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
			Name: opname,
			Opcode: opcode,
			Args: args,
		}

		ophash[opcode] = true
		cfg.operations[opname] = op
		log.Infof("Operation registered: %2X -> %s (%d)", op.Opcode, op.Name, len(op.Args))
	}
}

func (cfg *OpConfig) next() []string {
	if cfg.scanner.Scan() {
		cfg.line += 1
		return strings.Fields(strings.SplitN(cfg.scanner.Text(), "//",2)[0])
	}
	return nil
}


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
		err = errors.New(fmt.Sprintf("Unknown argument type `%s`", str))
	}
	return
}

func (cfg *OpConfig) Sprintf(format string, args ...interface{}) string {
	vars := append([]interface{}{cfg.fileName, cfg.line}, args...)
	return fmt.Sprintf("%s:%d > "+format, vars...)
}

func (cfg *OpConfig) Errorf(format string, args ...interface{}) {
	cfg.failed = true
	log.Errorf(cfg.Sprintf(format, args...))
}
