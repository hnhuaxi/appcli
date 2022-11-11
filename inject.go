package appcli

import (
	"fmt"
	"go/types"
	"io"
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

func Injects(funcs Map) {
	_InjectObjects = funcs
}

func (app *appImpl) GenerateInjects() error {
	var packages []string
	for _, inject := range app.Injects {
		packages = append(packages, inject.Package)
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
						alias: "",
						typ:   T,
					})
				case *types.Struct:
					names, _ := structsByType.At(T).([]string)
					names = append(names, name)
					structsByType.Set(T, names)
					nameTypes = append(nameTypes, &typeInfo{
						pkg:   object.Pkg(),
						name:  name,
						alias: "",
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
							alias: "",
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
							alias: "",
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
						alias: "",
						typ:   T,
					})
				default:
					names, _ := othersByType.At(T).([]string)
					names = append(names, name)
					othersByType.Set(T, names)
					others = append(others, &typeInfo{
						pkg:   object.Pkg(),
						name:  name,
						alias: "",
						typ:   T,
					})
				}

				if object.Pkg() != nil {
					pkgPaths[object.Pkg().Path()] = append(pkgPaths[object.Pkg().Path()], &typeInfo{
						pkg:   object.Pkg(),
						name:  name,
						alias: "",
						typ:   T,
					})
				}
			}
		}

	}

	genfile := File(app.Geninject)

	defer genfile.Close()
	return app.generateInjectgo(genfile, nameTypes, funcs, structs, others, pkgPaths)
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

func printMap(prefix string, m typeutil.Map) {
	var lines []string
	m.Iterate(func(T types.Type, names interface{}) {
		lines = append(lines, fmt.Sprintf("%s   %s", names, T))
	})
	sort.Strings(lines)
	for _, line := range lines {
		fmt.Printf(prefix, line)
	}
}

func (ctx *InjectContext) Imports() string {
	var (
		aliaies = make(map[string][]string)
	)

	for pkgName, typs := range ctx.pkgPaths {
		for _, typ := range typs {
			if typ.alias != "" {
				aliaies[pkgName] = append(aliaies[pkgName], typ.alias)
			}
		}
	}

	var lines []string
	for pkgName := range ctx.pkgPaths {
		lines = append(lines, Quote(pkgName))
		for _, alias := range aliaies[pkgName] {
			lines = append(lines, fmt.Sprintf("%s %s", alias, Quote(pkgName)))
		}
	}

	sort.Strings(lines)

	return strings.Join(lines, "\n")
}

func (ctx *InjectContext) Injects() string {
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

	for _, typ := range ctx.structs {
		if _, ok := idents[typ.name]; !ok {
			idents[typ.name] = true
			lines = append(lines, fmt.Sprintf("\t%s: %s,", Quote(typ.name), Quote(className(typ.name))))
		}
	}

	//
	for _, typ := range ctx.names {
		if _, ok := idents[typ.name]; !ok {
			idents[typ.name] = true
			lines = append(lines, fmt.Sprintf("\t%s: %s,", Quote(globalName(typ.name)), pkgName(typ.pkg, typ.name)))
		}
	}

	sort.Strings(lines)

	return strings.Join(lines, "\n")
}

func (ctx *InjectContext) InitCode() string {
	var classLines []string
	for _, typ := range ctx.structs {
		className := Quote(className(typ.name))
		classLines = append(classLines, fmt.Sprintf("\tappcli.RegisterCreate(%s, %s{})", className, pkgName(typ.pkg, typ.name)))
	}

	return "appcli.Injects(Injects)\n" +
		strings.Join(classLines, "\n")
}

func pkgName(pkg *types.Package, name string) string {
	return pkg.Name() + "." + name
}

var (
	globalVarnames = make(map[string]int)
)

func unique(varname string) string {
	var (
		tick int
	)

	defer func() {
		globalVarnames[varname] = tick
	}()

	tick, ok := globalVarnames[varname]
	if !ok {
		return varname
	}

	exists := func(s string) bool {
		_, ok := globalVarnames[varname]
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
	return unique("$" + camelCase(name))
}
