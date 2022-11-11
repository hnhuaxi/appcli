package appcli

type File string

func (f File) Read(p []byte) (n int, err error) {
	rawf, err := global.Open(string(f))
	if err != nil {
		return 0, err
	}

	return rawf.Read(p)
}

func (f File) Write(p []byte) (n int, err error) {
	rawf, err := global.OpenWrite(string(f))
	if err != nil {
		return 0, err
	}

	return rawf.Write(p)
}

func (f File) Close() (err error) {
	rawf, err := global.Open(string(f))
	if err != nil {
		return err
	}

	return rawf.Close()
}
