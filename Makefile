# Code generation commands
#
# Common usage:
#   make api
#   make rpc
#   make rpc RPC_PROTO=./proto/order.proto
#   make swagger
#   make swagger-swag
#   make gen

GOCTL ?= goctl
SWAG ?= swag

GO_MODULE ?= go-micro
STYLE ?= gozero

# goctl api input. The project documentation mentions api/gateway.api,
# but this repo currently does not contain an api directory.
API_FILE ?= ./api/gateway.api
API_OUT ?= ./internal/generated/api

PROTO_DIR ?= ./proto
PROTO_FILES := $(wildcard $(PROTO_DIR)/*.proto)
PROTO_TARGETS := $(patsubst $(PROTO_DIR)/%.proto,rpc-%,$(PROTO_FILES))
RPC_PROTO ?=
RPC_OUT ?= ./
RPC_ZRPC_OUT ?= ./internal/generated/rpc

SWAGGER_OUT ?= ./docs/swagger
SWAGGER_NAME ?= swagger.json
SWAGGER_MAIN ?= ./cmd/gateway-api/main.go

define RPC_CMD
$(GOCTL) rpc protoc $(1) --go_out=$(RPC_OUT) --go_opt=module=$(GO_MODULE) --go-grpc_out=$(RPC_OUT) --go-grpc_opt=module=$(GO_MODULE) --zrpc_out=$(RPC_ZRPC_OUT)/$(basename $(notdir $(1))) --style=$(STYLE)
endef

.PHONY: api rpc rpc-all rpc-one $(PROTO_TARGETS) swagger swagger-goctl swagger-swag gen gen-goctl help

api:
	$(if $(wildcard $(API_FILE)),,$(error API file not found: $(API_FILE). Create it or pass API_FILE=path/to/file.api))
	$(GOCTL) api go --api $(API_FILE) --dir $(API_OUT) --style=$(STYLE)

rpc:
ifeq ($(strip $(RPC_PROTO)),)
	$(MAKE) rpc-all
else
	$(MAKE) rpc-one RPC_PROTO=$(RPC_PROTO)
endif

rpc-all: $(PROTO_TARGETS)

$(PROTO_TARGETS): rpc-%:
	$(call RPC_CMD,$(PROTO_DIR)/$*.proto)

rpc-one:
	$(if $(strip $(RPC_PROTO)),,$(error RPC_PROTO is required, e.g. make rpc RPC_PROTO=./proto/order.proto))
	$(if $(wildcard $(RPC_PROTO)),,$(error RPC proto not found: $(RPC_PROTO)))
	$(call RPC_CMD,$(RPC_PROTO))

swagger: swagger-goctl

swagger-goctl:
	$(if $(wildcard $(API_FILE)),,$(error API file not found: $(API_FILE). goctl swagger generation requires a .api file))
	$(GOCTL) api plugin --api $(API_FILE) --dir $(SWAGGER_OUT) --plugin goctl-swagger="swagger -filename $(SWAGGER_NAME)" --style=$(STYLE)

# Current project workflow: Gin annotations + swaggo.
swagger-swag:
	$(SWAG) init -g $(SWAGGER_MAIN) -o $(SWAGGER_OUT)

# Full goctl generation flow. Requires API_FILE to exist.
gen-goctl: api rpc swagger-goctl

# Practical flow for the current repository layout.
gen: rpc swagger-swag

help:
	@echo "make api        Generate go-zero API code from API_FILE (default: ./api/gateway.api)"
	@echo "make rpc        Generate all proto/*.proto with goctl rpc protoc"
	@echo "make rpc RPC_PROTO=./proto/order.proto"
	@echo "make swagger    Generate Swagger from API_FILE with goctl-swagger"
	@echo "make swagger-swag Generate Swagger from Gin annotations with swag"
	@echo "make gen        Generate RPC code and current-project Swagger docs"
	@echo "make gen-goctl  Generate API, RPC, and Swagger through goctl"
	@echo "Override: API_FILE API_OUT PROTO_DIR RPC_PROTO RPC_OUT RPC_ZRPC_OUT SWAGGER_OUT SWAGGER_NAME STYLE"
