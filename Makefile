# goctl generation commands
# Usage examples:
#   make api API_FILE=./api/gateway.api API_OUT=./
#   make rpc                      # scan and generate all proto/*.proto
#   make rpc RPC_PROTO=./proto/order.proto
#   make swagger API_FILE=./api/gateway.api SWAGGER_OUT=./docs/swagger SWAGGER_NAME=swagger.json

GOCTL ?= goctl
GO_MODULE ?= go-micro

API_FILE ?= ./api/gateway.api
API_OUT ?= ./

PROTO_DIR ?= ./proto
PROTO_FILES := $(wildcard $(PROTO_DIR)/*.proto)
RPC_PROTO ?=
RPC_OUT ?= ./

SWAGGER_OUT ?= ./docs/swagger
SWAGGER_NAME ?= swagger.json

RPC_CMD = $(GOCTL) rpc protoc $1 --go_out=$(RPC_OUT) --go_opt=module=$(GO_MODULE) --go-grpc_out=$(RPC_OUT) --go-grpc_opt=module=$(GO_MODULE) --zrpc_out=$(RPC_OUT)

.PHONY: api rpc rpc-all rpc-one swagger swagger-swag gen help

api:
	$(GOCTL) api go -api $(API_FILE) -dir $(API_OUT)

rpc:
ifeq ($(strip $(RPC_PROTO)),)
	$(MAKE) rpc-all
else
	$(MAKE) rpc-one RPC_PROTO=$(RPC_PROTO)
endif

rpc-all:
	@echo "scan proto files: $(PROTO_FILES)"
	$(foreach f,$(PROTO_FILES),$(call RPC_CMD,$(f));)

rpc-one:
	$(call RPC_CMD,$(RPC_PROTO))

swagger:
	$(GOCTL) api plugin -api $(API_FILE) -dir $(SWAGGER_OUT) -plugin goctl-swagger="swagger -filename $(SWAGGER_NAME)"

# Keep existing swagger workflow in this project (swag + gin-swagger)
swagger-swag:
	swag init -g cmd/gateway-api/main.go -o ./docs/swagger

gen: api rpc swagger

help:
	@echo "make api      # 生成 API 代码（goctl api go）"
	@echo "make rpc      # 扫描并批量生成 proto/*.proto（默认输出到模块内对应目录，如 proto/*pb）"
	@echo "make rpc RPC_PROTO=./proto/order.proto # 生成单个 proto"
	@echo "make swagger  # 通过 goctl plugin 生成 swagger 文件"
	@echo "make swagger-swag # 使用 swag 生成当前项目 swagger 文档"
	@echo "可通过 API_FILE/API_OUT/PROTO_DIR/RPC_PROTO/RPC_OUT/SWAGGER_OUT/SWAGGER_NAME 覆盖默认值"