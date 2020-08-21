build:
	go build -o bin/pixel

linux:
	GOOS=linux go build -o bin/pixel

windows:
	GOOS=window GOARCH=386 go build -o bin/pixel.exe
