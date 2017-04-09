package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"git.practool.xyz/nova/goJASM/ijvmasm"
	"git.practool.xyz/nova/goJASM/opconf"
	"github.com/op/go-logging"
	"os"
)

var log *logging.Logger

func init() {
	log = logging.MustGetLogger("main")
	format := logging.MustStringFormatter(`%{module:10.10s} [%{color}%{level:.4s}%{color:reset}] %{message}`)
	logging.SetFormatter(format)
	logging.SetBackend(logging.NewLogBackend(os.Stdout, "", 0))
	logging.SetLevel(logging.NOTICE, "")
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
