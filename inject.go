package xtracego

import (
	"fmt"
	"io"
	"math/rand/v2"

	"github.com/Jumpaku/xtracego/internal"
)

type injector struct {
	cfg internal.Config
}

func NewInjector() *injector {
	return &injector{
		cfg: internal.Config{
			ResolveType: internal.ResolveType_CommandLineArguments,
		},
	}
}

func NewInjectorWithModule(moduleName string) *injector {
	return &injector{
		cfg: internal.Config{
			ModuleName:  moduleName,
			ResolveType: internal.ResolveType_PackageDirectory_Module,
		},
	}
}

func (i *injector) WithUniqueString(uniqueString string) *injector {
	if uniqueString == "" {
		alphabet := "abcdefghijklmnopqrstuvwxyz"
		v := []byte{}
		for i := 0; i < 8; i++ {
			v = append(v, alphabet[rand.IntN(len(alphabet))])
		}
		uniqueString = string(v)
	}
	i.cfg.UniqueString = uniqueString
	return i
}

func (i *injector) WithTraceStmt(traceStmt bool) *injector {
	i.cfg.TraceStmt = traceStmt
	return i
}

func (i *injector) WithTraceVar(traceVar bool) *injector {
	i.cfg.TraceVar = traceVar
	return i
}

func (i *injector) WithTraceCall(traceCall bool) *injector {
	i.cfg.TraceCall = traceCall
	return i
}

func (i *injector) WithShowGoroutine(showGoroutine bool) *injector {
	i.cfg.ShowGoroutine = showGoroutine
	return i
}

func (i *injector) WithShowTimestamp(showTimestamp bool) *injector {
	i.cfg.ShowTimestamp = showTimestamp
	return i
}

func (i *injector) WithLineWidth(lineWidth int) *injector {
	i.cfg.LineWidth = lineWidth
	return i
}

func (i *injector) InjectXtrace(src io.Reader, dst io.Writer) (err error) {
	srcBytes, err := io.ReadAll(src)
	if err != nil {
		return err
	}
	dstBytes, err := internal.ProcessCode(i.cfg, "", srcBytes)
	if err != nil {
		return err
	}
	if _, err := dst.Write(dstBytes); err != nil {
		return err
	}
	return nil
}

func (i *injector) GenerateLogger(dst io.Writer) (err error) {
	return internal.GetLibraryCode(i.cfg.LibraryPackageName(), i.cfg.UniqueString, dst)
}

func (i *injector) GenerateGoMod(dst io.Writer) (err error) {
	_, err = dst.Write([]byte(fmt.Sprintf(`module xtracego_tmp_%s`, i.cfg.UniqueString)))
	return err
}
