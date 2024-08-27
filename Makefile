BUILD_DIR ?= bin
BUILD_CLIENT_PACKAGE ?= ./pkg/client/
BUILD_SERVER_PACKAGE ?= ./cmd/
LOCAL_BIN:=$(CURDIR)/bin
BINARY_CLIENT_NAME = chat_client
BINARY_SERVER_NAME = chat_server

LDFLAGS += -s -w

build:
	go build -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/${BINARY_CLIENT_NAME} ${BUILD_CLIENT_PACKAGE}
	go build -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/${BINARY_SERVER_NAME} ${BUILD_SERVER_PACKAGE}