TARGET := server

all:
	go build -o $(TARGET) cmd/main.go
	air

run:
	air
