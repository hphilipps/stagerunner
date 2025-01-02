.PHONY: test build server demo

# Build the binary
build:
	go build -o stagerunner cmd/*.go

# Run tests
test:
	go test -v ./...

# Start the server
server:
	./stagerunner server --fail-probability 0.1 --workers 4

# Run demo commands
demo:
	bash demo.sh