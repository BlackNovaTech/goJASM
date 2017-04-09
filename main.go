package main

import (
	"git.practool.xyz/nova/goJASM/ijvmasm"
	"git.practool.xyz/nova/goJASM/opconf"
	"github.com/op/go-logging"
	"os"
	"flag"
	"strings"
	"path"
)

var log *logging.Logger
var flagInfo bool
var flagDebug bool
var flagConfig string
var flagOutput string
var flagForce bool

func init() {
	flag.BoolVar(&flagInfo, "info", false, "enable info message logging (default false)")
	flag.BoolVar(&flagDebug, "debug", false, "enable debug message logging (default false)")
	flag.StringVar(&flagConfig, "config", "", "specify custom ijvm configuration file")
	flag.StringVar(&flagOutput, "output", "", "specify output file. (default ./{filename}.ijvm)")
	flag.BoolVar(&flagForce, "force", false, "ignore most error messages and just yolo through (default false)")

	flag.Parse()

	log = logging.MustGetLogger("main")
	format := logging.MustStringFormatter(`%{module:10.10s} [%{color}%{level:.4s}%{color:reset}] %{message}`)
	logging.SetFormatter(format)
	logging.SetBackend(logging.NewLogBackend(os.Stdout, "", 0))
	if flagDebug {
		logging.SetLevel(logging.DEBUG, "")
	} else if flagInfo {
		logging.SetLevel(logging.INFO, "")
	} else {
		logging.SetLevel(logging.NOTICE, "")
	}
}

func main() {
	args := flag.Args()
	if len(args) == 0 {
		log.Critical("Please specify a file to compile")
		os.Exit(1)
	}

	var config *opconf.OpConfig
	if flagConfig == "" {
		config = opconf.NewDefaultOpConfig()
	} else {
		config = opconf.NewOpConfigFromPath(flagConfig)
	}

	input := args[0]
	output := flagOutput
	if output == "" {
		output = path.Base(strings.TrimSuffix(input, path.Ext(input))+".ijvm")
	}

	asm := ijvmasm.NewAssembler(input, config)
	err := asm.Parse()

	if err != nil && !flagForce {
		log.Critical("Assembly prematurely aborted:")
		log.Critical(err.Error())
		os.Exit(1)
	}

	if asm.Failed && !flagForce {
		log.Critical("Assembly failed")
		os.Exit(1)
	}

	out, err := os.Create(output)
	if err != nil {
		log.Critical("Could not open outputFile file:")
		log.Critical(err.Error())
		os.Exit(1)
	}
	defer out.Close()

	err = asm.Generate(out)
	if err != nil {
		log.Error(err.Error())
	}
	log.Notice("Assembly complete")
}
