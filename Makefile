MAKEFLAGS += --silent
SHELL := /usr/bin/env bash
PROJECT_NAME := github.com/carlosonunez/status
.PHONY: test unit e2e init tidy component _component_%

usage: ## Prints this help text.
	printf "make [target]\n\
Hack on $$(make project_name).\n\
\n\
TARGETS\n\
\n\
$$(fgrep -h '##' $(MAKEFILE_LIST) | fgrep -v '?=' | fgrep -v grep | sed 's/\\$$//' | sed -e 's/##//' | sed 's/^/  /g')\n\
\n\
ENVIRONMENT VARIABLES\n\
\n\
$$(fgrep '?=' $(MAKEFILE_LIST) | grep -v grep | sed 's/\?=.*##//' | sed 's/^/  /g')\n\
\n\
NOTES\n\
\n\
	- Adding a new stage? Add a comment with two pound signs after the stage name to add it to this help text.\n"

test: unit component e2e # Runs all tests

unit: _verify_ginkgo
unit: # Runs unit tests.
	ginkgo --label-filter '!e2e && !integration && !component' ./...

integration: _verify_ginkgo
integration: # Runs integration tests
	ginkgo --label-filter 'integration' ./...

e2e: _verify_ginkgo
e2e: # Runs e2e feature suites.
	ginkgo --label-filter tests/...

component: _verify_ginkgo _component_pub_sub

init: _verify_go
init:
	test -f go.mod || go mod init "$(PROJECT_NAME)"; \
	$(MAKE) tidy

tidy: _verify_go
	go mod tidy

_verify_go:
	if ! which go &>/dev/null; \
	then \
		>&2 echo "ERROR: Golang is not installed. Please install it first."; \
		exit 1; \
	fi; \
	exit 0

_verify_ginkgo:
	if ! which ginkgo &>/dev/null; \
	then \
		>&2 echo "ERROR: ginkgo not installed; run go get github.com/onsi/gingko/v2/ginkgo to do so"; \
		exit 1; \
	fi

_component_%:
	component=$$(echo "$@" | sed 's/^_component_//'); \
	>&2 echo "===> Running tests against third-party component: $$component"; \
	ginkgo --label-filter 'component' "third_party/$$component/..."
