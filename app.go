package appcli

import (
	"io"
	"os"

	"github.com/creasty/defaults"
	"gopkg.in/yaml.v3"
)

var (
	Version = "v0.0.1"
)

type App struct {
	Version string
	impl    appImpl
}

func (app *App) Generate() error {
	// inject external module
	return app.impl.GenerateInjects()
}

func (app *App) Execute(args []string) error {

	if err := app.Build(File("app.yaml")); err != nil {
		return err
	}

	return app.Run(args)
}

func (app *App) Build(rd io.Reader) error {

	dec := yaml.NewDecoder(rd)
	err := dec.Decode(&app.impl)
	if err != nil {
		return err
	}

	// app.InjectFuncs = _InjectFuncs
	_ = defaults.Set(&app.impl)

	// build actions
	return nil
}

func (app *App) Run(args []string) error {
	return app.impl.Run(args)
}

func (app *appImpl) writeOutput(val interface{}) error {
	enc := app.Output.Format.NewEncode(os.Stdout)
	return enc.Encode(val)
}
