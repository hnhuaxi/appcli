// Code generated by apicli. DO NOT EDIT.

package main

import (
	"github.com/hnhuaxi/appcli"
	"github.com/hnhuaxi/appcli/env"
)

var Injects = map[string]interface{}{
	"map":     appcli.Map,
	"version": appcli.Version,
}

var InjectFuncs = map[string]interface{}{
	"actionFunc":      appcli.ActionFunc,
	"compileProgram":  appcli.CompileProgram,
	"loadPackages":    appcli.LoadPackages,
	"newObject":       appcli.NewObject,
	"newPlainEncoder": appcli.NewPlainEncoder,
	"quote":           appcli.Quote,
	"registerCreate":  appcli.RegisterCreate,
}

var InjectTypes = map[string]interface{}{
	"App":           appcli.App{},
	"Command":       appcli.Command{},
	"Flag":          appcli.Flag{},
	"Inject":        appcli.Inject{},
	"InjectContext": appcli.InjectContext{},
	"Match":         appcli.Match{},
	"Output":        appcli.Output{},
}

func init() {
	env.Injects(Injects)
	env.InjectTypes(InjectTypes)
	env.InjectFuncs(InjectFuncs)

}
