package expr

import (
	"strings"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"github.com/hnhuaxi/appcli/program"
	"github.com/hnhuaxi/appcli/utils"
)

type Program struct {
	progs []*vm.Program
	local map[string]interface{}
}

func Compile(exp string, env interface{}) (*Program, error) {
	var (
		progrm = Program{
			progs: make([]*vm.Program, 0),
			local: make(map[string]interface{}),
		}
		lines = strings.Split(exp, "\n")
	)

	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			var (
				prog *vm.Program
				err  error
			)
			if menv, ok := env.(map[string]interface{}); ok {

				prog, err = expr.Compile(line, expr.Env(utils.Merge(menv, utils.ResultCtx(nil))))
				if err != nil {
					return nil, err
				}
			} else {
				prog, err = expr.Compile(line, expr.Env(env))
				if err != nil {
					return nil, err
				}
			}
			progrm.progs = append(progrm.progs, prog)
		}
	}

	return &progrm, nil
}

func (progr *Program) Run(env interface{}) (output interface{}, err error) {
	var prevResult interface{}

	withResult := func(fn func(_env interface{}) (interface{}, error)) (output interface{}, err error) {
		if menv, ok := env.(map[string]interface{}); ok {
			if prevResult != nil {
				menv["$_"] = prevResult
			}
			defer func() {
				prevResult = output
			}()

			output, err = fn(menv)
			return
		} else {
			return fn(env)
		}
	}

	for _, prog := range progr.progs {
		output, err = withResult(func(env interface{}) (interface{}, error) {
			output, err = expr.Run(prog, env)
			if err != nil {
				return nil, err
			}
			return output, nil
		})

		if err != nil {
			return nil, err
		}
	}

	return output, nil
}

func init() {
	program.Register(&Program{}, func(exp string, env interface{}) (program.Program, error) {
		return Compile(exp, env)
	})
}
