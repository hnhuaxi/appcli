package anko

import (
	"errors"

	ienv "github.com/hnhuaxi/appcli/env"
	"github.com/hnhuaxi/appcli/program"
	"github.com/mattn/anko/ast"
	"github.com/mattn/anko/env"
	_ "github.com/mattn/anko/packages"
	"github.com/mattn/anko/parser"
	"github.com/mattn/anko/vm"
)

type Program struct {
	stmt ast.Stmt
}

func Compile(exp string, env interface{}) (*Program, error) {
	stmt, err := parser.ParseSrc(exp)
	if err != nil {
		return nil, err
	}

	return &Program{
		stmt: stmt,
	}, nil
}

func (prog *Program) Run(envs interface{}) (output interface{}, err error) {
	e := env.NewEnv()
	menv, ok := envs.(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid env type, must map[string]interface{}")
	}

	// for

	for key, val := range menv {
		if err := e.Define(key, val); err != nil {
			return nil, err
		}
	}

	for key, val := range ienv.AllInjectFuncs {
		if err := e.Define(key, val); err != nil {
			return nil, err
		}
	}

	for key, val := range ienv.AllInjectTypes {
		if err := e.DefineGlobalType(key, val); err != nil {
			return nil, err
		}
	}

	for key, val := range ienv.BuiltinFuncs {
		if err := e.Define(key, val); err != nil {
			return nil, err
		}
	}

	for key, val := range ienv.AllGlobalOjects {
		if err := e.DefineGlobal(key, val); err != nil {
			return nil, err
		}
	}

	for key, val := range ienv.AllInjectObjects {
		if err := e.Define(key, val); err != nil {
			return nil, err
		}
	}

	return vm.Run(e, nil, prog.stmt)
}

func init() {
	program.Register(&Program{}, func(exp string, env interface{}) (program.Program, error) {
		return Compile(exp, env)
	})
}

var _ program.Program = &Program{}
