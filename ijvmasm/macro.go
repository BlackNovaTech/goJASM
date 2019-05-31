package ijvmasm

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

func (asm *Assembler) executeMacro(method *Method, line string) {
	if strings.HasPrefix(line, "#print") {
		param := strings.TrimSpace(strings.TrimPrefix(line, "#print"))
		if param == "" {
			asm.Errorf("#print called without arguments")
			return
		}

		textToPrint, err := strconv.Unquote(param)
		if err != nil {
			asm.Errorf("error unquoting #print param `%s`: %+v", param, err)
		}

		logrus.WithField("text", textToPrint).Infof("[.%s] Evaluating macro #print", method.name)

		for _, char := range textToPrint {
			asm.parseInstruction(method, fmt.Sprintf("BIPUSH %d", char))
			asm.parseInstruction(method, "OUT")
		}

	}
}
