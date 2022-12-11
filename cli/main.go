package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hnhuaxi/appcli"
	"github.com/hnhuaxi/appcli/internal"
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
				Name:    "quite",
				Value:   true,
				Aliases: []string{"q"},
				Usage:   "dismiss output",
			},
			&cli.StringFlag{
				Name:  "init",
				Usage: "initial a project, include main.go, go.mod files",
			},
			&cli.StringFlag{
				Name:  "pkg",
				Usage: "package name",
			},
			&cli.StringFlag{
				Name:  "app-version",
				Value: "v0.0.1",
				Usage: "specify app version",
			},
			&cli.StringFlag{
				Name:  "app-name",
				Usage: "specify app name",
			},
			&cli.StringFlag{
				Name:  "app-usage",
				Value: "Demo for yaml config cli app",
				Usage: "a app usage",
			},
			&cli.StringFlag{
				Name:  "app-description",
				Value: "Description for yaml config cli app",
				Usage: "a app description",
			},
			&cli.StringFlag{
				Name:  "app-author",
				Usage: "devlopment author",
			},
		},
		Action: func(ctx *cli.Context) error {
			var app = &appcli.App{}

			if ctx.String("init") == "" {
				if err := app.Build(appcli.File("app.yaml")); err != nil {
					log.Fatalf("build app.yaml config error %s", err)
				}

				if err := app.Generate(); err != nil {
					log.Fatalf("generate file error %s", err)
				}

				if !ctx.Bool("quite") {
					fmt.Printf("copy below code to create 'main.go'\n---\n%s\n", mainCode)
				}
			} else {
				ar := internal.Compile(&internal.ProjectContext{
					Name:        ctx.String("app-name"),
					PkgName:     ctx.String("pkg"),
					Version:     ctx.String("app-version"),
					Author:      ctx.String("app-author"),
					Usage:       ctx.String("app-usage"),
					Description: ctx.String("app-description"),
					License:     ctx.String("app-license"),
				})

				if err := internal.CopyArchive(ctx.String("init"), ar); err != nil {
					log.Fatalf("copy arichive error %s", err)
				}

				if err := app.Build(appcli.File("app.yaml")); err != nil {
					log.Fatalf("build app.yaml config error %s", err)
				}

				if err := app.Generate(); err != nil {
					log.Fatalf("generate file error %s", err)
				}
			}

			return nil
		},
	}).Run(os.Args)
}
