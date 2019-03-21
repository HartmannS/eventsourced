.PHONY: help clean build test test-no-race test-stress travis mocks run sniff format docker

include ./VERSION

BIN_FILE := $(PROJECT)
EXP_PORT := 2069

# ---

GIT_BRANCH := $(shell git name-rev --name-only HEAD | sed "s/~.*//")
GIT_COMMIT := $(shell git describe --always)
GIT_AUTHOR := $(shell git config user.email)
GIT_ORIGIN := $(shell git config --get remote.origin.url | sed s/\.git$$//)
GIT_STATUS :=
IMG_STATUS := clean

ifneq (, $(shell git status -s))
	STATUS_SEP   := -
	GIT_STATUS := tainted
	IMG_STATUS := $(GIT_STATUS)
endif

NOW := $(shell date "+%Y%m%d%H%M%S")
BUILDER_IMG := $(PROJECT)/builder

DOCKER_TAG    := $(BIN_FILE)
DOCKER_TAGS   := --tag $(DOCKER_TAG):$(VERSION)-$(GIT_COMMIT)-$(NOW)$(STATUS_SEP)$(GIT_STATUS)
DOCKER_TAGS   += --tag $(DOCKER_TAG):latest

DOCKER_LABELS := --label branch="$(GIT_BRANCH)"
DOCKER_LABELS += --label commit="$(GIT_COMMIT)"
DOCKER_LABELS += --label author="$(GIT_AUTHOR)"
DOCKER_LABELS += --label origin="$(GIT_ORIGIN)"
DOCKER_LABELS += --label status="$(IMG_STATUS)"

# ---

CONF_FILE := config.yml

HAS_LOCAL_CONF := 0
ifneq ("", "$(wildcard ./$(CONF_FILE))")
	HAS_LOCAL_CONF := 1
endif

COVERAGE_DIR := coverage
COVERAGE_MKD := mkdir -p ./$(COVERAGE_DIR) 2> /dev/null
COVERAGE_OUT := $(COVERAGE_DIR)/cover.out

MODULES := $(shell go list ./... | grep -v "mock\|/conf\|stress\|cmd" | grep -v "^$(BIN_FILE)$$")
DEBRIS  := (\.prof|\.test\|.out\|.log)$$

TEST_PARAMS := -count=1 -v -covermode=atomic -coverprofile=./$(COVERAGE_OUT) $(MODULES)
TEST_REPORT := go tool cover -html=./$(COVERAGE_OUT) -o ./$(COVERAGE_DIR)/index.html && echo "report	<file://$(PWD)/$(COVERAGE_DIR)/index.html>"

TEST_DORACE := $(COVERAGE_MKD); CGO_ENABLED=1 go test $(TEST_PARAMS) -race
TEST_NORACE := $(COVERAGE_MKD); CGO_ENABLED=0 go test $(TEST_PARAMS)
TEST_STRESS := $(COVERAGE_MKD); CGO_ENABLED=0 go test -count=1 -v eventsourced/intern/stress

GCFLAGS :=
LDFLAGS := -s -w
LDFLAGS += -X main.project=$(PROJECT)
LDFLAGS += -X main.version=$(VERSION)
LDFLAGS += -X main.commit=$(GIT_COMMIT)
LDFLAGS += -X main.status=$(GIT_STATUS)
LDFLAGS := -ldflags "$(LDFLAGS)"

# ---

help:                   # Displays this list
	@cat Makefile \
	    | grep "^[a-z][a-z0-9_-]\+:" \
	    | sed -r "s/:[^#]*?#?(.*)?/\r\t\t\t-\1/" \
	    | sed "s/^/ • make /"

clean:                  # Removes build/test artifacts
	@rm -f ./$(BIN_FILE) 2> /dev/null
	@rm -rf ./$(COVERAGE_DIR) 2> /dev/null
	@find . -type f | grep -E "$(DEBRIS)" | xargs -I{} rm {};

build: clean            # Builds binary
	@CGO_ENABLED=0 go build $(GCFLAGS) $(LDFLAGS) $(ARGS) ./cmd/eventsourced.go

run: build              # Builds and executes binary
	@./$(BIN_FILE)

test: clean             # Runs tests with coverage, with -race
	@($(TEST_DORACE) | grep -v "PASS\|RUN\|^coverage") && $(TEST_REPORT)

test-no-race: clean     # Runs tests, gcc not required, no -race
	@($(TEST_NORACE) | grep -v "PASS\|RUN\|^coverage") && $(TEST_REPORT)

test-stress: clean      # Runs a self stress test
	@echo "Low ulimit -n (now `ulimit -n`) will lead to failures. Stressing ..."
	@($(TEST_STRESS) | grep -v "PASS\|RUN\|^coverage")

travis:                 # Travis CI target (see .travis.yml), runs tests
    ifndef TRAVIS
	    @echo "Fail: requires Travis runtime"
    else
	    @$(TEST_NORACE) && goveralls -coverprofile=./$(COVERAGE_OUT) -service=travis-ci
    endif

mocks:                  # Creates mocks for tests
	@find . -type f | grep \.go$$ | xargs -n1 go generate

sniff:                  # Linting & vetting (void on success)
	@echo "License header ..."
	@find . -type f -name \*.go | grep -v "intern/mock" | xargs grep -L "^//.*MIT"
	@echo "Running go fmt ..."
	@gofmt -d .
	@echo "Running go lint ..."
	@golint ./...
	@echo "Running go vet ..."
	@go vet ./...

format:                 # Formats (changes) Go source files
	@gofmt -w .

builder-image:
	@echo "Creating $(BUILDER_IMG) image ..."
	@docker build --tag $(BUILDER_IMG):latest ./build

docker:            		# Creates a tagged Docker image
    ifeq (,$(HAS_LOCAL_CONF))
		@touch ./$(CONF_FILE)
    endif

	@>/dev/null 2>&1 docker inspect --type=image $(BUILDER_IMG) || make -s builder-image
	@docker build --no-cache $(DOCKER_TAGS) $(DOCKER_LABELS) .

	@>/dev/null 2>&1 docker rmi `docker images -f "dangling=true" -q` "" || true
	@echo "Run container with: docker run -d -p $(EXP_PORT):$(EXP_PORT) $(DOCKER_TAG):latest"

    ifeq (,$(HAS_LOCAL_CONF))
		@touch ./$(CONF_FILE)
    endif
