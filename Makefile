.PHONY: build install clean test

BINARY_NAME=koko
INSTALL_DIR=$(HOME)/.local/bin

build:
	go build -o $(BINARY_NAME) main.go

install: build
	mkdir -p $(INSTALL_DIR)
	cp $(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Installed $(BINARY_NAME) to $(INSTALL_DIR)/$(BINARY_NAME)"

clean:
	rm -f $(BINARY_NAME) *.wav
