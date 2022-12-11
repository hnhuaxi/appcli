package internal

import (
	"bytes"
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/tools/txtar"
)

const ProjectFiles = `This a generate cli app scaffold project.
-- go.mod --
module {{ .PkgName }}

go 1.19
-- main.go --
package main

import (
	"os"

	"github.com/hnhuaxi/appcli"
)

var app = appcli.App{}

func main() {
	app.Execute(os.Args)
}
-- app.yaml --
---
name: {{ .Name }}
usage: {{ .Usage }}
version: {{ .Version }}
author: {{ .Author }}
description: {{ .Description }}
output:
  format: json
flags:
  - name: "hadoop"
    type: "Bool"
    usage: "use hadoop"
    value: true
commands:
  - name: "doo"
    usage: "do the doo"
    description: "no really"
    flags:
      - name: "flag"
        type: "Bool"
        value: true
    action: print(flag)
action: 'printf("hello\n")'
license:
  header: This file bleongs to clig
  copyright: Copyright © 2019 clig
  text: |
    Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

    The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

    THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
`

type ProjectContext struct {
	Name        string
	PkgName     string
	Version     string
	Author      string
	Usage       string
	Description string
	License     string
}

func Compile(ctx *ProjectContext) (ar *txtar.Archive) {
	var buf bytes.Buffer
	tmpl := template.Must(template.New("scaffold").Parse(ProjectFiles))
	if err := tmpl.Execute(&buf, ctx); err != nil {
		log.Fatalf("generate scaffold template error %s", err)
		return nil
	}

	ar = txtar.Parse(buf.Bytes())
	return
}

func CopyArchive(dest string, src *txtar.Archive) error {
	_ = os.MkdirAll(dest, os.ModeDir|os.ModePerm)

	for _, file := range src.Files {
		f, err := os.Create(filepath.Join(dest, file.Name))
		if err != nil {
			return err
		}

		if _, err := io.Copy(f, bytes.NewBuffer(file.Data)); err != nil {
			return err
		}

		f.Close()
	}

	return nil
}
