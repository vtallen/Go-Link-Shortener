TARGET := server

all: tls
	go build -o $(TARGET) cmd/
	air

tls: # Generates a testing certificate
	go run /usr/local/go/src/crypto/tls/generate_cert.go --host localhost

run:
	sudo setcap 'cap_net_bind_service=+ep' /tmp/main
	air

lines: # Shows how many lines of code are in the project
	cat views/* cmd/* pkg/codegen/* | wc -l

clean:
	rm -r -f tmp
	rm -r -f $(TARGET)
	rm -f cert.pem
	rm -f key.pem
