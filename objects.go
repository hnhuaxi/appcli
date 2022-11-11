package appcli

import (
	"reflect"

	"github.com/imdario/mergo"
)

type objectCtor func(attr Map) interface{}

var (
	objects = make(map[string]objectCtor)
)

func NewObject(className string, attrs ...Map) interface{} {
	var attr Map
	if len(attrs) > 0 {
		attr = attrs[0]
	}
	return objects[className](attr)
}

func RegisterCreate(className string, instance interface{}) {
	if _, ok := objects[className]; ok {
		panic("always register className")
	}

	objects[className] = func(attr Map) interface{} {
		var v = reflect.ValueOf(instance)
		v = reflect.Indirect(v)

		instance := reflect.New(v.Type()).Interface()
		if attr != nil {
			_ = mergo.Map(instance, attr)
		}
		return instance
	}
}

func className(name string) string {
	return "$$Class" + name
}
