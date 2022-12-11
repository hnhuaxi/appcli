package tengo

import (
	"errors"

	"github.com/d5/tengo/v2"
	"github.com/hnhuaxi/appcli/program"
)

type Program struct {
	compl *tengo.Compiled
}

func Compile(exp string, env interface{}) (*Program, error) {
	script := tengo.NewScript([]byte(exp))
	compl, err := script.Compile()
	if err != nil {
		return nil, err
	}

	return &Program{
		compl: compl,
	}, nil
}

func (progr *Program) Run(envs interface{}) (output interface{}, err error) {
	menv, ok := envs.(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid env type, must map[string]interface{}")
	}

	for key, val := range menv {
		if err := progr.compl.Set(key, val); err != nil {
			return nil, err
		}
	}

	err = progr.compl.Run()
	return nil, err
}

func init() {
	program.Register(&Program{}, func(exp string, env interface{}) (program.Program, error) {
		return Compile(exp, env)
	})
}
