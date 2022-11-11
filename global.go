package appcli

import "os"

type Global map[string]*os.File

var global = make(Global)

func (g *Global) Open(filename string) (f *os.File, err error) {
	defer func() {
		(*g)[filename] = f
	}()

	if _f, ok := (*g)[filename]; ok {
		return _f, nil
	}

	f, err = os.Open(filename)
	return
}

func (g *Global) OpenWrite(filename string) (f *os.File, err error) {
	defer func() {
		(*g)[filename] = f
	}()

	if _f, ok := (*g)[filename]; ok {
		return _f, nil
	}

	f, err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	return
}
