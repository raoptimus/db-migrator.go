SHELL = /bin/bash -e
BASEDIR=$(shell pwd)
GIT_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null)
GIT_COMMIT ?= $(shell git rev-parse --short HEAD)
GIT_TAG ?= $(shell git describe --tags --abbrev=0 HEAD 2>/dev/null || false)
VERSION ?= $(shell [[ "${GIT_TAG}" == "" ]] && echo 0.0.1 || echo ${GIT_TAG})
export ${VERSION}
LDFLAGS=-ldflags "-s -w -X main.Version=${GIT_TAG} -X main.GitCommit=${GIT_COMMIT}"
COVERAGE_DIR ?= .coverage
BUILD_DIR ?= .build
SOURCE_FILES ?= ./...
TEST_PATTERN ?= .
TEST_OPTIONS ?=
PKG_NAME = db-migrator
APTLY_BASE_URL ?=
APTLY_REPO_MASTER ?= xenial
APTLY_DIST ?= xenial
APTLY_REPO ?= xenial
APTLY_PREFIX ?= $(shell [[ ${GIT_BRANCH} == "master" ]] && echo "stable" || echo "develop")
PACKAGE_FILE = "$(PKG_NAME)-$(VERSION).deb"
PKG_WORKDIR = "${BUILD_DIR}/${PKG_NAME}-${VERSION}"
DOCKER_ID_USER = raoptimus
DOCKER_IMAGE = "${PKG_NAME}"

help:
	@echo "VERSION: ${VERSION}"
	@echo "GIT_BRANCH: ${GIT_BRANCH}"
	@echo "GIT_TAG: ${GIT_TAG}"
	@echo "GIT_COMMIT: ${GIT_COMMIT}"

build-docker-image:
	@docker build \
		--build-arg VERSION=${VERSION} \
		--build-arg GIT_BRANCH=${GIT_BRANCH} \
		--build-arg GIT_COMMIT=${GIT_COMMIT} \
		--build-arg GIT_TAG=${GIT_TAG} \
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

test:
	@[ -d ${COVERAGE_DIR} ] || mkdir -p ${COVERAGE_DIR}
	@go test $(TEST_OPTIONS) \
		-failfast \
		-race \
		-coverpkg=./... \
		-covermode=atomic \
		-coverprofile=${COVERAGE_DIR}/coverage.txt $(SOURCE_FILES) \
		-run $(TEST_PATTERN) \
		-timeout=2m

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

