SHELL = /bin/bash -e
.DEFAULT_GOAL=help
BASEDIR=$(shell pwd)
GIT_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null)
GIT_COMMIT ?= $(shell git rev-parse --short HEAD)
GIT_TAG ?= $(shell git describe --tags --abbrev=0 HEAD 2>/dev/null || echo "v0.0.0")
VERSION = $(shell echo "${GIT_TAG}" | grep -Eo "v([0-9]+\.[0-9]+\.[0-9]+)" | cut -dv -f2)
export ${VERSION}
LDFLAGS=-ldflags "-s -w -X main.Version=${GIT_TAG} -X main.GitCommit=${GIT_COMMIT}"
REPORT_DIR ?= .report
BUILD_DIR ?= .build
PKG_NAME = db-migrator
APTLY_BASE_URL ?=
APTLY_REPO_MASTER ?= xenial
APTLY_DIST ?= xenial
APTLY_REPO ?= xenial
APTLY_PREFIX ?= $(shell [[ ${GIT_BRANCH} == "master" ]] && echo "stable" || echo "develop")
PACKAGE_FILE = "$(PKG_NAME)-$(VERSION).deb"
PKG_WORKDIR = "${BUILD_DIR}/${PKG_NAME}-${VERSION}"
DOCKER_ID_USER = raoptimus
DOCKER_PASS ?= ""
DOCKER_IMAGE = "${PKG_NAME}"
export GO_IMAGE_VERSION="1.22"

help: ## Show help message
	@cat $(MAKEFILE_LIST) | grep -e "^[a-zA-Z_\-]*: *.*## *" | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

version:
	@echo "VERSION: ${VERSION}"
	@echo "GIT_BRANCH: ${GIT_BRANCH}"
	@echo "GIT_TAG: ${GIT_TAG}"
	@echo "GIT_COMMIT: ${GIT_COMMIT}"

build-docker-image: ## Build docker image
	@#docker login -u "${DOCKER_ID_USER}" -p "${DOCKER_PASS}" docker.io
	@docker build \
		--platform linux/x86_64 \
		--build-arg VERSION=${VERSION} \
		--build-arg GIT_BRANCH=${GIT_BRANCH} \
		--build-arg GIT_COMMIT=${GIT_COMMIT} \
		--build-arg GIT_TAG=${GIT_TAG} \
		--build-arg GO_IMAGE_VERSION=${GO_IMAGE_VERSION} \
		-f ./docker/image/build/Dockerfile \
		-t ${DOCKER_ID_USER}/${DOCKER_IMAGE}:"${VERSION}-alpine" \
		-t ${DOCKER_ID_USER}/${DOCKER_IMAGE}:latest  ./

	@docker push ${DOCKER_ID_USER}/${DOCKER_IMAGE}:"${VERSION}-alpine"
	@docker push ${DOCKER_ID_USER}/${DOCKER_IMAGE}:latest

build: help
	@[ -d ${BUILD_DIR} ] || mkdir -p ${BUILD_DIR}
	CGO_ENABLED=0 go build ${LDFLAGS} -o ${BUILD_DIR}/${PKG_NAME} cmd/${PKG_NAME}/main.go
	@file  ${BUILD_DIR}/${PKG_NAME}
	@du -h ${BUILD_DIR}/${PKG_NAME}

test-coverage: ## Coverage
	@[ -d ${REPORT_DIR} ] || mkdir -p ${REPORT_DIR}
	@-go test \
          -cover \
          -covermode=atomic \
          -coverprofile=${REPORT_DIR}/coverage.txt \
          ./... \
          -timeout=2m \
          -v

test-unit: ## Run only Unit tests
	@go test $$(go list ./... | grep -v mock) \
		-buildvcs=false \
		-failfast \
		-short \
		-timeout=2m \
		-v

test-integration: ## Run Integration Tests only
	@go test $$(go list ./... | grep -v mock) \
		-buildvcs=false \
		-failfast \
		-run Integration \
		-tags=integration \
		-v

build-deb: build
	@echo "deb package $(PKG_NAME) building..."
	@rm -rf ${BUILD_DIR}/${PKG_NAME}-*

	@mkdir -p ${PKG_WORKDIR}/DEBIAN
	@cp -r scripts/debian/* ${PKG_WORKDIR}/DEBIAN/

	@sed -i ${PKG_WORKDIR}/DEBIAN/control -e "s/<VERSION>/${VERSION}/" ${PKG_WORKDIR}/DEBIAN/control
	@sed -i ${PKG_WORKDIR}/DEBIAN/control -e "s/<PACKAGE>/${PKG_NAME}/" ${PKG_WORKDIR}/DEBIAN/control

	@mkdir -p ${PKG_WORKDIR}/opt/bin
	@cp ${BUILD_DIR}/${PKG_NAME} ${PKG_WORKDIR}/opt/bin/${PKG_NAME}

	@dpkg-deb --build $(PKG_WORKDIR) $(PKG_WORKDIR).deb
	@rm -rf ${PKG_WORKDIR}

publish-deb:
	# upload and publish
	@curl --fail --connect-timeout 5 -X POST -F file=@${BUILD_DIR}/${PACKAGE_FILE} ${APTLY_BASE_URL}/api/files/debian
	@curl --fail --connect-timeout 5 -X POST ${APTLY_BASE_URL}/api/repos/${APTLY_REPO}/file/debian/${PACKAGE_FILE}
	@curl --fail --connect-timeout 5 -X PUT ${APTLY_BASE_URL}/api/publish/filesystem:ci:${APTLY_PREFIX}/${APTLY_DIST}

lint: ## Run linter
	@[ -d ${REPORT_DIR} ] || mkdir -p ${REPORT_DIR}
	golangci-lint run --timeout 5m

install-mockery:
	@mockery --version &> /dev/null || go install github.com/vektra/mockery/v2@latest
	@mockery --version

gen-mocks: install-mockery ## Run mockery
	@bash -c 'for d in $$(find . ! -path "**/mock*" ! -path "**/.*" -name "**.go" -exec dirname {} \; | sort --unique); do \
	if [[ "$$d" ==  "." ]]; then continue; fi; \
	pkg=mock$$(basename -- "$${d/./''}"); \
	mockery --srcpkg=$${d} --outpkg=$${pkg} --output=$${d}/$${pkg} --all --with-expecter=true; \
	done; \
	'

gen-mocks-dry-run: install-mockery ## Run mockery --dry-run=true
	@bash -c 'for d in $$(find . ! -path "**/mock*" ! -path "**/.*" -name "**.go" -exec dirname {} \; | sort --unique); do \
	if [[ "$$d" ==  "." ]]; then continue; fi; \
	mockery --srcpkg=$${d} --output=$${d}/mock$$(basename -- "$${d/./''}") --all --dry-run=true; \
	done; \
	'

start:
	@docker-compose -f "docker-compose.yml" -f "docker-compose.dev.yml" up -d
