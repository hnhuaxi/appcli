package utils

import "github.com/imdario/mergo"

type Map = map[string]interface{}

func Merge(envs ...Map) Map {
	if len(envs) < 1 {
		return envs[0]
	}

	var dst = make(Map)
	for _, env := range envs {
		_ = mergo.MapWithOverwrite(&dst, env)
	}
	return dst
}

func ResultCtx(result interface{}) Map {
	return Map{"$_": result}
}
