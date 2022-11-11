package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hnhuaxi/appcli"
	"github.com/urfave/cli/v2"
)

const mainCode = `package main

import (
	"log"
	"os"

	"github.com/hnhuaxi/appcli"
)

var app = appcli.App{}

func main() {
	app.Execute(os.Args)
}`

func main() {

	(&cli.App{
		Name: "appcli",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "quite",
				Value: true,
			},
		},
		Action: func(ctx *cli.Context) error {
			var app = &appcli.App{}
			if err := app.Build(appcli.File("app.yaml")); err != nil {
				log.Fatalf("build app.yaml config error %s", err)
			}

			if err := app.Generate(); err != nil {
				log.Fatalf("generate file error %s", err)
			}

			if !ctx.Bool("quite") {
				fmt.Printf("copy below code to create 'main.go'\n---\n%s\n", mainCode)
			}

			return nil
		},
	}).Run(os.Args)
}
