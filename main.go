package main

import (
	"io"
	"os"

	"fmt"

	"github.com/BlackNovaTech/gojasm/ijvmasm"
	"github.com/BlackNovaTech/gojasm/opconf"
	"github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
)

var (
	flagInfo     bool
	flagDebug    bool
	flagConfig   string
	flagOutput   string
	flagForce    bool
	flagAutoWide bool
	flagSymbols  bool
	flagVersion  bool
)

// Linker Variables
var (
	Version   = "DEVELOPMENT"
	BuildDate = "<dev>"
)

func init() {
	flag.BoolVarP(&flagInfo, "info", "i", false, "enable info message logging")
	flag.BoolVarP(&flagDebug, "debug", "d", false, "enable debug message logging")
	flag.StringVarP(&flagConfig, "config", "c", "", "specify custom ijvm configuration file")
	flag.StringVarP(&flagOutput, "output", "o", "out.ijvm", "specify output file. (default out.ijvm)")
	flag.BoolVarP(&flagForce, "force", "f", false, "ignore most error messages and just yolo through")
	flag.BoolVarP(&flagAutoWide, "widen", "w", false, "automatically add WIDE operations when required")
	flag.BoolVarP(&flagSymbols, "symbols", "s", false, "generate symbol blocks")
	flag.BoolVarP(&flagVersion, "version", "v", false, "output version information")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s inputfile\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if flagVersion {
		printVersion()
		os.Exit(0)
	}

	if flagDebug {
		logrus.SetLevel(logrus.DebugLevel)
	} else if flagInfo {
		logrus.SetLevel(logrus.InfoLevel)
	} else {
		logrus.SetLevel(logrus.WarnLevel)
	}
}

func main() {
	args := flag.Args()
	if len(args) == 0 {
		logrus.Fatal("Please specify a file to compile")
	}

	var config *opconf.OpConfig
	if flagConfig == "" {
		config = opconf.NewDefaultOpConfig()
	} else {
		config = opconf.NewOpConfigFromPath(flagConfig)
	}

	input := args[0]
	output := flagOutput

	asm := ijvmasm.NewAssembler(input, config)
	asm.AutoWide = flagAutoWide
	ok, err := asm.Parse()

	if err != nil && !flagForce {
		logrus.WithError(err).Fatal("Assembly prematurely aborted:")
	}

	if !ok && !flagForce {
		logrus.Fatalln("Assembly failed")
	}

	var out io.Writer = os.Stdout
	if output != "-" {
		outf, err := os.Create(output)
		if err != nil {
			logrus.WithError(err).Fatal("Could not open output file")
		}
		defer outf.Close()
		out = outf
	}

	if err = asm.Generate(out); err != nil {
		logrus.WithError(err).Error("Error generating bytecode")
	}

	if flagSymbols {
		logrus.Info("Generating Symbols...")
		if err = asm.GenerateDebugSymbols(out); err != nil {
			logrus.WithError(err).Error("Error generating symbols")
		}
	}
}

func printVersion() {
	fmt.Printf("gojasm version %s\n", Version)
	fmt.Printf("Built at: %s\n", BuildDate)
}
