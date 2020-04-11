.PHONY: bins clean
PROJECT_ROOT = github.com/rppala90/apdDemo

export PATH := $(GOPATH)/bin:$(PATH)

# default target
default: clean

apdServer:
	go build -i -o bin/apdserver ./cmd/apd/apdserver/*.go

apdDemo:
	go build -i -o bin/apddemo ./cmd/apd/*.go

bins: apdServer \
	apdDemo \

clean:
	rm -rf bin
