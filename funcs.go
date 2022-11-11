package appcli

import "fmt"

var BuiltinObjects = Map{
	"print":  fmt.Print,
	"printf": fmt.Printf,
	"new":    NewObject,
	"global": RegisterGlobal,
}

var (
	_InjectObjects = make(Map)
	_GlobalObjects = make(Map)
)

func RegisterGlobal(name string, instance interface{}) error {
	_GlobalObjects[name] = instance
	return nil
}
