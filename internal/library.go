package internal

import (
	"fmt"
	"io"
	"text/template"
)

var xtraceGo string = `
package {{.PackageName}}

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

func getTimestamp() string {
	return time.Now().In(time.UTC).Format(time.RFC3339)
}

func getGoroutineId() string {
	return string(bytes.Split(debug.Stack(), []byte(" "))[1])
}

func getFuncName(stack int) string {
	pc, _, _, _ := runtime.Caller(stack)
	name := runtime.FuncForPC(pc).Name()
	vs := strings.Split(name, "/")
	return vs[len(vs)-1]
}

func getPrefix(stack int, showTimestamp, showGoroutine bool) string {
	prefix := ""
	if showTimestamp {
		prefix += fmt.Sprintf("%20s ", getTimestamp())
	}
	if showGoroutine {
		prefix += fmt.Sprintf("[%2s] ", getGoroutineId())
	}
	return prefix + getFuncName(stack) + ": "
}

func PrintlnStatement_{{.UniqueString}}(stack int, width int, line, source string, showTimestamp, showGoroutine bool) {
	prefix := getPrefix(stack, showTimestamp, showGoroutine)
	lenPrefix, lenLine, lenSource := len(prefix), len(line), len(source)
	dots := ""
	if lenPrefix+lenLine+lenSource < width {
		dots = strings.Repeat("-", width-lenPrefix-lenLine-lenSource)
	}
	_, _ = fmt.Fprintln(os.Stderr, prefix + line + dots + source)
}

func PrintlnVariable_{{.UniqueString}}(stack int, width int, varName string, varValue any, showTimestamp, showGoroutine bool) {
	prefix := getPrefix(stack, showTimestamp, showGoroutine)
	variable := fmt.Sprintf("[VAR] %s=%#v", varName, varValue)
	lenPrefix, lenVariable := len(prefix), len(variable)
	if lenPrefix+lenVariable >= width {
		_, _ = fmt.Fprintln(os.Stderr, (prefix + variable)[:width-3] + "...")
	} else {
		_, _ = fmt.Fprintln(os.Stderr, (prefix + variable))
	}
}

func PrintlnReturnVariable_{{.UniqueString}}(stack int, width int, varName string, varValue any, showTimestamp, showGoroutine bool) {
	prefix := getPrefix(stack, showTimestamp, showGoroutine)
	variable := fmt.Sprintf("[VAR] %s=%#v", varName, varValue)
	lenPrefix, lenVariable := len(prefix), len(variable)
	if lenPrefix+lenVariable >= width {
		_, _ = fmt.Fprintln(os.Stderr, (prefix + variable)[:width-4] + " ...")
	} else {
		_, _ = fmt.Fprintln(os.Stderr, (prefix + variable))
	}
}

func PrintlnCall_{{.UniqueString}}(width int, signature string, showTimestamp, showGoroutine bool) {
	prefix := getPrefix(3, showTimestamp, showGoroutine)
	callStr := prefix + "[CALL] " + signature
	if len(callStr) >= width {
		callStr = callStr[:width-4] + " ..."
	}
	_, _ = fmt.Fprintln(os.Stderr, callStr)
}

func PrintlnReturn_{{.UniqueString}}(width int, signature string, showTimestamp, showGoroutine bool) {
	prefix := getPrefix(3, showTimestamp, showGoroutine)
	returnStr := prefix + "[RETURN] " + signature
	if len(returnStr) >= width {
		returnStr = returnStr[:width-4] + " ..."
	}
	_, _ = fmt.Fprintln(os.Stderr, returnStr)
}
`

var xtraceGoTemplate = template.Must(template.New("xtrace.go.tpl").Parse(xtraceGo))

type XtraceGoData struct {
	PackageName  string
	UniqueString string
}

func GetLibraryCode(packageName, uniqueString string, w io.Writer) (err error) {
	d := XtraceGoData{PackageName: packageName, UniqueString: uniqueString}
	if err := xtraceGoTemplate.Execute(w, d); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}
	return nil
}
