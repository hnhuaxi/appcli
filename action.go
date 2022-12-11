package appcli

import (
	"reflect"

	"github.com/antonmedv/expr"
	"github.com/hnhuaxi/appcli/program"
	"github.com/urfave/cli/v2"
)

type Action string

type ActionFunc = func(*cli.Context) error

type (
	// Program = vm.Program
	Env = expr.Option
)

var (
	compileCaches = make(map[string]program.Program)
)

func buildAction(exp string, inenv interface{}) (prog program.Program, err error) {
	var ok bool
	defer func() {
		compileCaches[exp] = prog
	}()

	if prog, ok = compileCaches[exp]; !ok {
		prog, err = CompileProgram(exp, inenv)
	}
	return
}

func allFlagNames(ctx *cli.Context) []string {
	var names []string
	for _, c := range ctx.Lineage() {
		if c.Command == nil {
			continue
		}

		for _, f := range c.Command.Flags {
			names = append(names, f.Names()...)
		}
	}

	if ctx.App != nil {
		for _, f := range ctx.App.Flags {
			names = append(names, f.Names()...)
		}
	}

	return dedup(names)
}

func buildCtxEnv(ctx *cli.Context) Map {
	var (
		env   = make(Map)
		ctxM  = make(Map)
		names = allFlagNames(ctx)
	)

	env["ctx"] = ctxM
	for _, name := range names {
		val := ctx.Value(name)
		switch x := val.(type) {
		case cli.Timestamp:
			ctxM[name] = x.Value()
		case cli.StringSlice:
			ctxM[name] = x.Value()
		default:
			ctxM[name] = ctx.Value(name)
		}
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

func (act Action) Compile(env interface{}) (program.Program, error) {
	return buildAction(string(act), env)
}
