package appcli

import (
	"errors"

	"github.com/hnhuaxi/appcli/program"
	"github.com/hnhuaxi/appcli/program/anko"
	_ "github.com/hnhuaxi/appcli/program/expr"
)

var (
	DefaultCompiler program.Program = &anko.Program{}
)

func CompileProgram(exp string, env interface{}) (program.Program, error) {
	compiler, ok := program.Lookup(DefaultCompiler)
	if !ok {
		return nil, errors.New("invalid compiler")
	}

	return compiler(exp, env)
}
