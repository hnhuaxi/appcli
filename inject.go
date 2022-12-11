package appcli

import (
	"bytes"
	"fmt"
	"go/types"
	"io"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"unicode"

	_ "embed"

	"github.com/creasty/defaults"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/types/typeutil"
)

var (
	//go:embed inject.gen.tmpl
	injectgenTmpl string
)

type Inject struct {
	Package string
	Alias   string
	Methods []Match
	Objects []Match
	Action  Action
}

type Match struct {
	Regexp   string
	Prefix   string
	Contains []string
	Excludes []string
}

var (
	globalAliaies = make(map[string]string)
)

func LoadPackages(pkgNames ...string) ([]*packages.Package, error) {
	cfg := &packages.Config{Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedImports | packages.NeedDeps | packages.NeedExportFile | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedTypesSizes}
	pkgs, err := packages.Load(cfg, pkgNames...)
	if err != nil {
		return nil, err
	}
	if packages.PrintErrors(pkgs) > 0 {
		return nil, err
	}

	return pkgs, nil
}

func (app *appImpl) GenerateInjects() error {
	var (
		packages []string
	)

	globalAliaies = make(map[string]string)

	for _, inject := range app.Injects {
		packages = append(packages, inject.Package)
		if inject.Alias != "" {
			globalAliaies[inject.Package] = inject.Alias
		}
	}

	getAlias := func(pkg *types.Package) string {
		return globalAliaies[pkg.Path()]
	}

	pkgs, err := LoadPackages(packages...)
	if err != nil {
		return err
	}

	var (
		namesByType   typeutil.Map // value is []string
		funcsByType   typeutil.Map // value is []string
		structsByType typeutil.Map
		othersByType  typeutil.Map
		funcs         []*typeInfo
		nameTypes     []*typeInfo
		structs       []*typeInfo
		others        []*typeInfo
		pkgPaths      = make(map[string][]*typeInfo)
	)

	for _, pkg := range pkgs {
		scope := pkg.Types.Scope()
		for _, name := range scope.Names() {
			if unicode.IsUpper(([]rune)(name)[0]) {
				var (
					object = scope.Lookup(name)
					T      = object.Type()
				)
				switch T.(type) {
				case *types.Signature:
					names, _ := funcsByType.At(T).([]string)
					names = append(names, name)
					funcsByType.Set(T, names)
					funcs = append(funcs, &typeInfo{
						pkg:   object.Pkg(),
						name:  name,
						alias: getAlias(object.Pkg()),
						typ:   T,
					})
				case *types.Struct:
					names, _ := structsByType.At(T).([]string)
					names = append(names, name)
					structsByType.Set(T, names)
					nameTypes = append(nameTypes, &typeInfo{
						pkg:   object.Pkg(),
						name:  name,
						alias: getAlias(object.Pkg()),
						typ:   T,
					})
				case *types.Named:
					switch T.Underlying().(type) {
					case *types.Struct:
						_ = T
						names, _ := structsByType.At(T).([]string)
						names = append(names, name)
						structsByType.Set(T, names)
						// mset := types.NewMethodSet(types.NewPointer(T))
						// for i := 0; i < mset.Len(); i++ {
						// 	fmt.Println(mset.At(i).Obj().Name(), mset.At(i))
						// }
						structs = append(structs, &typeInfo{
							pkg:   object.Pkg(),
							name:  name,
							alias: getAlias(object.Pkg()),
							typ:   T,
						})
					case *types.Signature:
						// names, _ := namesByType.At(T).([]string)
						// names = append(names, name)
						// namesByType.Set(T, names)
					default:
						names, _ := othersByType.At(T).([]string)
						names = append(names, name)
						othersByType.Set(T, names)
						others = append(others, &typeInfo{
							pkg:   object.Pkg(),
							name:  name,
							alias: getAlias(object.Pkg()),
							typ:   T,
						})
					}
				case *types.Map, *types.Basic:
					names, _ := namesByType.At(T).([]string)
					names = append(names, name)
					namesByType.Set(T, names)
					nameTypes = append(nameTypes, &typeInfo{
						pkg:   object.Pkg(),
						name:  name,
						alias: getAlias(object.Pkg()),
						typ:   T,
					})
				default:
					names, _ := othersByType.At(T).([]string)
					names = append(names, name)
					othersByType.Set(T, names)
					others = append(others, &typeInfo{
						pkg:   object.Pkg(),
						name:  name,
						alias: getAlias(object.Pkg()),
						typ:   T,
					})
				}

				if object.Pkg() != nil {
					pkgPaths[object.Pkg().Path()] = append(pkgPaths[object.Pkg().Path()], &typeInfo{
						pkg:   object.Pkg(),
						name:  name,
						alias: getAlias(object.Pkg()),
						typ:   T,
					})
				}
			}
		}
	}

	var (
		genfile = File(app.Geninject)
		buf     bytes.Buffer
	)

	defer genfile.Close()

	if err := app.generateInjectgo(&buf, nameTypes, funcs, structs, others, pkgPaths); err != nil {
		return err
	}

	return app.importsFormat(genfile, &buf)
}

func typeName(typ *types.TypeName) *types.Package {
	if typ != nil {
		return typ.Pkg()
	}
	return nil
}

type typeInfo struct {
	pkg   *types.Package
	name  string
	alias string
	typ   types.Type
}

type InjectContext struct {
	PkgName  string `default:"main"`
	names    []*typeInfo
	funcs    []*typeInfo
	structs  []*typeInfo
	others   []*typeInfo
	pkgPaths map[string][]*typeInfo
}

func (app *appImpl) generateInjectgo(w io.Writer, names, funcs, structs, others []*typeInfo, pkgPaths map[string][]*typeInfo) error {
	tmpl := template.Must(template.New("inject.go").Parse(injectgenTmpl))
	var ctx = InjectContext{
		names:    names,
		funcs:    funcs,
		structs:  structs,
		others:   others,
		pkgPaths: pkgPaths,
	}
	_ = defaults.Set(&ctx)
	return tmpl.Execute(w, &ctx)
}

func (app *appImpl) importsFormat(w io.Writer, source io.Reader) error {
	// 调用 goimports 格式化代码
	cmd := exec.Command("goimports")
	cmd.Stdin = source
	cmd.Stdout = w
	return cmd.Run()

}

func (ctx *InjectContext) Imports() string {
	var (
		aliaies = make(map[string]string)
	)

	for pkgName, typs := range ctx.pkgPaths {
		for _, typ := range typs {
			if typ.alias != "" {
				aliaies[pkgName] = typ.alias
			}
		}
	}

	var lines = []string{
		// Quote("github.com/hnhuaxi/appcli"),
		Quote("github.com/hnhuaxi/appcli/env"),
	}

	for pkgName := range ctx.pkgPaths {
		if alias, ok := aliaies[pkgName]; ok {
			lines = append(lines, fmt.Sprintf("\t%s %s", alias, Quote(pkgName)))
		} else {
			lines = append(lines, Quote(pkgName))
		}
	}

	lines = dedup(lines)
	sort.Strings(lines)

	return strings.Join(lines, "\n")
}

func (ctx *InjectContext) Injects() string {
	var (
		idents = make(map[string]bool)
		lines  []string
	)

	for _, typ := range ctx.names {
		if _, ok := idents[typ.name]; !ok {
			idents[typ.name] = true
			lines = append(lines, fmt.Sprintf("\t%s: %s,", Quote(globalName(typ.name)), pkgName(typ.pkg, typ.name)))
		}
	}

	sort.Strings(lines)
	lines = dedup(lines)

	return strings.Join(lines, "\n")
}

func (ctx *InjectContext) InjectTypes() string {
	var (
		idents = make(map[string]bool)
		lines  []string
	)

	for _, typ := range ctx.structs {
		if _, ok := idents[typ.name]; !ok {
			idents[typ.name] = true
			lines = append(lines, fmt.Sprintf("\t%s: %s,", Quote(typ.name), varStruct(typ.pkg, typ.name)))
		}
	}

	sort.Strings(lines)
	lines = dedup(lines)

	return strings.Join(lines, "\n")
}

func (ctx *InjectContext) InjectFuncs() string {
	var (
		idents = make(map[string]bool)
		lines  []string
	)

	for _, typ := range ctx.funcs {
		if _, ok := idents[typ.name]; !ok {
			idents[typ.name] = true
			lines = append(lines, fmt.Sprintf("\t%s: %s,", Quote(globalName(typ.name)), pkgName(typ.pkg, typ.name)))
		}
	}

	sort.Strings(lines)
	lines = dedup(lines)

	return strings.Join(lines, "\n")
}

func (ctx *InjectContext) InitCode() string {
	var classLines []string

	return "env.Injects(Injects)\n" +
		"env.InjectTypes(InjectTypes)\n" +
		"env.InjectFuncs(InjectFuncs)\n" +
		strings.Join(classLines, "\n")
}

func pkgName(pkg *types.Package, name string) string {
	var (
		pkgName = pkg.Name()
		alias   = globalAliaies[pkg.Path()]
	)

	if alias != "" {
		pkgName = alias
	}
	return pkgName + "." + name
}

func varStruct(pkg *types.Package, name string) string {
	return pkgName(pkg, name) + "{}"
}

var (
	globalVars = make(map[string]int)
)

func unique(varname string) string {
	var (
		tick int
	)

	defer func() {
		globalVars[varname] = tick
	}()

	tick, ok := globalVars[varname]
	if !ok {
		return varname
	}

	exists := func(s string) bool {
		_, ok := globalVars[varname]
		return ok
	}

	for j := tick; j < tick+5; j++ {
		if !exists(varname + strconv.Itoa(j)) {
			tick = j
			break
		}
	}

	varname += strconv.Itoa(tick)
	return varname
}

func varName(name string) string {
	return unique(camelCase(name))
}

func globalName(name string) string {
	// return unique("$" + camelCase(name))
	return unique(camelCase(name))

}

func dedup[T string | int](sliceList []T) []T {
	allKeys := make(map[T]bool)
	list := []T{}
	for _, item := range sliceList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}
