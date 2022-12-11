package appcli

import (
	"fmt"
	"go/types"
	"sort"
	"strconv"

	"github.com/segmentio/go-camelcase"
	"golang.org/x/tools/go/types/typeutil"
)

func Quote(s string) string {
	return strconv.Quote(s)
}

var q = Quote

func camelCase(s string) string {
	return camelcase.Camelcase(s)
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
