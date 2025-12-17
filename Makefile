BIN_DIR := $(HOME)/go/bin
BIN_NAME := codestash

.PHONY: build install

build:
	go build ./...

install:
	go build -o $(BIN_DIR)/$(BIN_NAME) .
