export COMPOSE_DOCKER_CLI_BUILD = 1
export DOCKER_BUILDKIT = 1
export COMPOSE_PROJECT_NAME = ledger
export ARCH = $(shell uname -m | sed 's/x86_64/amd64/' | sed 's/aarch64/arm64/')
export META = $(shell git rev-parse --abbrev-ref HEAD 2> /dev/null | sed 's:.*/::')
export VERSION = $(shell git fetch --tags --force 2> /dev/null; tags=($$(git tag --sort=-v:refname)) && ([ "$${\#tags[@]}" -eq 0 ] && echo v0.0.0 || echo $${tags[0]}) | sed -e "s/^v//")

.ONESHELL:
.PHONY: arm64
.PHONY: amd64

.PHONY: all
all: bootstrap sync test package bbtest

.PHONY: package
package:
	@$(MAKE) bundle-binaries-$(ARCH)
	@$(MAKE) bundle-debian-$(ARCH)
	@$(MAKE) bundle-docker-$(ARCH)

.PHONY: bundle-binaries-%
bundle-binaries-%: %
	@\
		docker \
		compose \
		run \
		--rm package \
		--arch linux/$^ \
		--source /go/src/github.com/jancajthaml-openbank/ledger-rest \
		--output /project/packaging/bin
	@\
		docker \
		compose \
		run \
		--rm package \
		--arch linux/$^ \
		--source /go/src/github.com/jancajthaml-openbank/ledger-unit \
		--output /project/packaging/bin

.PHONY: bundle-debian-%
bundle-debian-%: %
	@\
		docker \
		compose \
		run \
		--rm debian-package \
		--version $(VERSION) \
		--arch $^ \
		--pkg ledger \
		--source /project/packaging

.PHONY: bundle-docker-%
bundle-docker-%: %
	@\
		docker \
		build \
		-t openbank/ledger:$^-$(VERSION).$(META) \
		-f packaging/docker/$^/Dockerfile \
		.

.PHONY: bootstrap
bootstrap:
	@\
		docker \
		compose \
		build \
		--force-rm go

.PHONY: lint
lint:
	@\
		docker \
		compose \
		run \
		--rm lint \
		--source /go/src/github.com/jancajthaml-openbank/ledger-rest \
	|| :
	@\
		docker \
		compose \
		run \
		--rm lint \
		--source /go/src/github.com/jancajthaml-openbank/ledger-unit \
	|| :

.PHONY: sec
sec:
	@\
		docker \
		compose \
		run \
		--rm sec \
		--source /go/src/github.com/jancajthaml-openbank/ledger-rest \
	|| :
	@\
		docker \
		compose \
		run \
		--rm sec \
		--source /go/src/github.com/jancajthaml-openbank/ledger-unit \
	|| :

.PHONY: sync
sync:
	@\
		docker \
		compose \
		run \
		--rm sync \
		--source /go/src/github.com/jancajthaml-openbank/ledger-rest
	@\
		docker \
		compose \
		run \
		--rm sync \
		--source /go/src/github.com/jancajthaml-openbank/ledger-unit

.PHONY: scan-%
scan-%: %
	@\
		docker \
		scan \
		openbank/ledger:$^-$(VERSION).$(META) \
		--file ./packaging/docker/$^/Dockerfile \
		--exclude-base

.PHONY: test
test:
	@\
		docker \
		compose \
		run \
		--rm test \
		--source /go/src/github.com/jancajthaml-openbank/ledger-rest \
		--output /project/reports/unit-tests
	@\
		docker \
		compose \
		run \
		--rm test \
		--source /go/src/github.com/jancajthaml-openbank/ledger-unit \
		--output /project/reports/unit-tests

.PHONY: release
release:
	@\
		docker \
		compose \
		run \
		--rm release \
		--version $(VERSION) \
		--token ${GITHUB_RELEASE_TOKEN}

.PHONY: bbtest
bbtest:
	@docker compose up -d bbtest
	@docker exec -t $$(docker compose ps -q bbtest) python3 /opt/app/bbtest/main.py
	@docker compose down -v
