build:
	@go build -o bin/appcli ./cli

install:
	@rm -rf ~/go/bin/appcli
	@cp bin/appcli ~/go/bin/appcli