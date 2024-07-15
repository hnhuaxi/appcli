package env

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/gotidy/ptr"
	"github.com/hnhuaxi/utils/convert"
	"github.com/imdario/mergo"
	"github.com/kr/pretty"
)

type Map = map[string]interface{}

var (
	AllInjectObjects = make(Map)
	AllInjectTypes   = make(Map)
	AllInjectFuncs   = make(Map)
	AllGlobalOjects  = make(Map)
)

var BuiltinFuncs = Map{
	"print":         fmt.Print,
	"printf":        fmt.Printf,
	"println":       fmt.Println,
	"pprint":        pretty.Print,
	"pprintf":       pretty.Printf,
	"global":        RegisterGlobal,
	"set":           Set,
	"now":           Now,
	"ts":            Timestamp,
	"tsm":           TimestampMs,
	"createContext": context.Background,
	"withValue":     context.WithValue,
	"withCancel":    context.WithCancel,
	"rel":           Rel,
	"pstr":          ptr.String,
	"pint":          ptr.Int,
	"pint8":         ptr.Int8,
	"pint16":        ptr.Int16,
	"pint32":        ptr.Int32,
	"pint64":        ptr.Int64,
	"puint":         ptr.UInt,
	"puint8":        ptr.UInt8,
	"puint16":       ptr.UInt16,
	"puint32":       ptr.UInt32,
	"puint64":       ptr.UInt64,
	"pbool":         ptr.Bool,
	"pfloat32":      ptr.Float32,
	"pfloat64":      ptr.Float64,
	"pcomplex64":    ptr.Complex64,
	"pcomplex128":   ptr.Complex128,
	"pbyte":         ptr.Byte,
}

func Injects(objects Map) {
	AllInjectObjects = objects
}

func InjectFuncs(funcs Map) {
	AllInjectFuncs = funcs
}

func InjectTypes(types Map) {
	for typ, typVal := range types {
		var v = reflect.ValueOf(typVal)
		if v.Kind() == reflect.Struct {
			AllInjectTypes[typ] = reflect.New(reflect.TypeOf(typVal)).Interface()
		}
	}
}

func RegisterGlobal(name string, instance interface{}) error {
	AllGlobalOjects[name] = instance
	return nil
}

func Set(dest, src interface{}) error {
	if srcm, ok := src.(map[interface{}]interface{}); ok {
		return mergo.Map(dest, concM2M(srcm))
	} else {
		return mergo.Map(dest, src)
	}
}

type (
	AnyMap = map[interface{}]interface{}
	StrMap = map[string]interface{}
)

func concM2M(m AnyMap) StrMap {
	var mm = make(StrMap)
	for key, val := range m {

		if mmv, ok := val.(AnyMap); ok {
			mm[convert.ToStr(key)] = concM2M(mmv)
		} else {
			mm[convert.ToStr(key)] = val
		}
	}

	return mm
}

func Now() time.Time {
	return time.Now()
}

func Timestamp() int64 {
	return time.Now().Unix()
}

func TimestampMs() int64 {
	return time.Now().UnixMilli()
}

func Rel(v interface{}) interface{} {
	var vv = reflect.ValueOf(v)
	vv = reflect.Indirect(vv)
	return vv.Interface()
}
