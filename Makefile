include common.mk

# Check if Go's linkers flags are set in common.mk and add them as extra flags.
ifneq ($(GOLDFLAGS),)
	GO_EXTRA_FLAGS += -ldflags $(GOLDFLAGS)
endif

# Set all target as the default target.
all: build

# Build.
build:
	@$(ECHO) "$(MAGENTA)*** Building Go code...$(OFF)"
	@$(GO) build $(GOFLAGS) -o oasis-registry/oasis-registry ./oasis-registry

# Build examples.
build-examples:
	@$(ECHO) "$(MAGENTA)*** Building examples...$(OFF)"
	@cd examples/lookup && \
		$(GO) build

# Generate test vectors.
gen_vectors:
	@$(ECHO) "$(MAGENTA)*** Generating test vectors ...$(OFF)"
	@$(GO) run ./$@

# Format code.
fmt-targets := fmt-go fmt-sh

fmt-go:
	@$(ECHO) "$(CYAN)*** Running Go formatters...$(OFF)"
	@gofumpt -s -w .
	@gofumports -w -local github.com/oasisprotocol/metadata-registry-tools .

fmt-sh:
	@$(ECHO) "$(CYAN)*** Running Shell formatters...$(OFF)"
	@shfmt -l -w .

fmt: $(fmt-targets)

# Lint code, commits and documentation.
lint-targets := lint-go lint-sh lint-docs lint-git lint-go-mod-tidy

lint-go:
	@$(ECHO) "$(CYAN)*** Running Go linters...$(OFF)"
	@env -u GOPATH golangci-lint run

lint-sh:
	@$(ECHO) "$(CYAN)*** Running Shell linters...$(OFF)"
	@shfmt -d .

lint-git:
	@$(ECHO) "$(CYAN)*** Runnint gitlint...$(OFF)"
	@$(CHECK_GITLINT)

lint-docs:
	@$(ECHO) "$(CYAN)*** Runnint markdownlint-cli...$(OFF)"
	@npx markdownlint-cli '**/*.md'

lint-go-mod-tidy:
	@$(ECHO) "$(CYAN)*** Checking go mod tidy...$(OFF)"
	@$(ENSURE_GIT_CLEAN)
	@$(CHECK_GO_MOD_TIDY)

lint: $(lint-targets)

# Test.
test-targets := test-unit test-cli

test-unit:
	@$(ECHO) "$(CYAN)*** Running unit tests...$(OFF)"
	@$(GO) test -v -race ./...

test-cli: build
	@$(ECHO) "$(CYAN)*** Running CLI tests...$(OFF)"
	@./tests/test-cli-general.sh

test: $(test-targets)

# Clean.
clean:
	@$(ECHO) "$(CYAN)*** Cleaning up ...$(OFF)"
	@$(GO) clean -x

# List of targets that are not actual files.
.PHONY: \
	all build build-examples \
	gen_vectors \
	$(fmt-targets) fmt \
	$(lint-targets) lint \
	$(test-targets) test \
	clean
