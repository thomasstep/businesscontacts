GOCMD=go
GOBUILD=$(GOCMD) build
BINARY_NAME=contacts

build:
	$(GOBUILD) -o $(BINARY_NAME) -v
