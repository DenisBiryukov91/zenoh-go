# Define directories
SRC_DIR?=./examples
BIN_DIR?=./bin

# Go commands
GO_BUILD=go build
GO_FMT=go fmt
GO_CLEAN=go clean

# Find all example directories starting with "z_"
EXAMPLES := $(shell find $(SRC_DIR) -maxdepth 1 -type d -name 'z_*')

# Define binaries for each example
BINARIES := $(patsubst $(SRC_DIR)/%, %, $(EXAMPLES))

# Default target to build all binaries
.PHONY: all
all: $(BINARIES)

# Build binary for each example
$(BINARIES):
	@mkdir -p $(BIN_DIR)
	$(GO_BUILD) -o $(BIN_DIR)/$(notdir $@) $(SRC_DIR)/$@/$(notdir $@).go

# Format the source code
.PHONY: fmt
fmt:
	@output=$$($(GO_FMT) $(SRC_DIR)/...); \
	if [ -n "$$output" ]; then \
		echo "Code format was updated:"; \
		echo "$$output"; \
		exit 1; \
	fi

# Clean up all generated binaries
.PHONY: clean
clean:
	$(GO_CLEAN)
	rm -rf $(BIN_DIR)

# Install dependencies
.PHONY: deps
deps:
	go mod tidy

# Run tests 
.PHONY: test 
test:
	go test ./tests

# Run examples tests 
.PHONY: test_examples
test_examples:
	python3 tests/test_examples.py $(BIN_DIR)

# Run examples tests QUIC
.PHONY: test_examples_quic
test_examples_quic:
	python3 tests/test_examples.py -l "quic/localhost:7449" -e "quic/localhost:7449" -c tests/quic.json5 $(BIN_DIR)
