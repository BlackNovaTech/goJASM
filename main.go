package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"git.practool.xyz/nova/goJASM/ijvmasm"
	"git.practool.xyz/nova/goJASM/opconf"
	"github.com/op/go-logging"
	"os"
	"flag"
)

var log *logging.Logger
var logInfo bool
var logDebug bool

func init() {
	flag.BoolVar(&logInfo, "info", false, "enable info message logging (default false)")
	flag.BoolVar(&logDebug, "debug", false, "enable debug message logging (default false)")
	flag.Parse()

	log = logging.MustGetLogger("main")
	format := logging.MustStringFormatter(`%{module:10.10s} [%{color}%{level:.4s}%{color:reset}] %{message}`)
	logging.SetFormatter(format)
	logging.SetBackend(logging.NewLogBackend(os.Stdout, "", 0))
	if logDebug {
		logging.SetLevel(logging.DEBUG, "")
	} else if logInfo {
		logging.SetLevel(logging.INFO, "")
	} else {
		logging.SetLevel(logging.NOTICE, "")
	}
}

func main() {
	var err error
	config := opconf.NewOpConfig("stupid.conf")
	asm := ijvmasm.NewAssembler("test2.jas", config)
	err = asm.Parse()
	if err != nil {
		log.Critical("Assembly prematurely aborted:")
		log.Critical(err.Error())
		os.Exit(1)
	}
	if asm.Failed {
		log.Critical("Assembly failed")
		os.Exit(1)
	}

	buf := new(bytes.Buffer)
	err = asm.Generate(buf)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(hex.Dump(buf.Bytes()))
}
