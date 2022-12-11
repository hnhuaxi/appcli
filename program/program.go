package program

import "reflect"

type (
	typenode interface {
	}

	Program interface {
		Run(env interface{}) (output interface{}, err error)
	}

	Compiler func(exp string, env interface{}) (Program, error)
)

var (
	programs = make(map[reflect.Type]Compiler)
)

func Register(node typenode, compile Compiler) {
	typ := reflect.ValueOf(node).Type()
	if _, ok := programs[typ]; ok {
		panic("awayls register compiler")
	} else {
		programs[typ] = compile
	}
}

func Lookup(node typenode) (Compiler, bool) {
	typ := reflect.ValueOf(node).Type()
	compiler, ok := programs[typ]
	return compiler, ok
}
