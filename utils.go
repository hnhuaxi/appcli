package appcli

import (
	"strconv"

	"github.com/segmentio/go-camelcase"
)

func Quote(s string) string {
	return strconv.Quote(s)
}

func camelCase(s string) string {
	return camelcase.Camelcase(s)
}
