FLAGS = -ldflags '-extldflags "-static"'

all:
	go build $(FLAGS) -o bin/launch main.go
	GOOS=linux GOARCH=amd64 go build $(FLAGS) -o bin/launch-amd64 main.go

