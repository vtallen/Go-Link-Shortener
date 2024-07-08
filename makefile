TARGET := server

all: 
	cd cmd && go build -o ../$(TARGET) .	

deb: clean all
	./$(TARGET)

tls: # Generates a testing certificate
	go run /usr/local/go/src/crypto/tls/generate_cert.go --host localhost

run:
	sudo setcap 'cap_net_bind_service=+ep' /tmp/main
	air

lines: # Shows how many lines of code are in the project
	cat views/* cmd/* pkg/codegen/* internal/sessmngt/* internal/conf/* internal/sessmngt/* | wc -l

build:
	go build cmd/.

clean:
	rm -r -f tmp
	rm -r -f $(TARGET)
	rm -r -f shortener.db
	# rm -f cert.pem
	# rm -f key.pem
