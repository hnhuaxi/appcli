package appcli

import (
	"reflect"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"github.com/urfave/cli/v2"
)

type Action string

type ActionFunc = func(*cli.Context) error

type (
	Program = vm.Program
	Env     = expr.Option
)

var (
	compileCaches = make(map[string]*Program)
)

func buildAction(exp string, inenv interface{}) (prog *Program, err error) {
	var ok bool
	defer func() {
		compileCaches[exp] = prog
	}()

	if prog, ok = compileCaches[exp]; !ok {
		prog, err = expr.Compile(exp, expr.Env(inenv))
	}
	return
}

func buildCtxEnv(ctx *cli.Context) Map {
	var env = make(Map)

	var names = ctx.FlagNames()
	for _, name := range names {
		env[name] = ctx.Value(name)
	}

	return env
}

func buildFlagEnv(flags []*Flag) Map {
	var env = make(Map)
	for _, flag := range flags {
		name := getFieldString(flag.Flag, "Name")
		env[name] = getField(flag.Flag, "Value")
	}

	return env
}

func getField(val interface{}, field string) interface{} {
	var v = reflect.ValueOf(val)
	v = reflect.Indirect(v)
	fe := v.FieldByName(field)
	if !fe.IsValid() || fe.IsZero() {
		return nil
	}
	return fe.Interface()
}

func getFieldString(val interface{}, field string) string {
	var v = reflect.ValueOf(val)
	v = reflect.Indirect(v)
	fe := v.FieldByName(field)
	return fe.String()
}

func (act Action) Compile(env interface{}) (*Program, error) {
	return buildAction(string(act), env)
}
