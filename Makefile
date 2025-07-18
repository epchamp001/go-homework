# Путь к бинарнику
BINARY_NAME := pvz

# Путь к файлу линтера
LINT_CONFIG := .golangci.yaml

.PHONY: update linter build start run

update:
	go mod tidy

linter:
	golangci-lint run --config=$(LINT_CONFIG)

build:
	go build -o $(BINARY_NAME) cmd/pvz/main.go cmd/pvz/setup_flag.go

start:
	./$(BINARY_NAME)

# Обновить зависимости, запустить линтеры, собрать и запустить приложение
run: update linter build start

# Пути
LOCAL_BIN := $(CURDIR)/bin
VENDOR_DIR         := $(CURDIR)/vendor-proto
PROTO_TMPL_DIR     := $(VENDOR_DIR)/protobuf
TYPE_TMPL_DIR      := $(VENDOR_DIR)/googleapis
OPENAPI_TMPL_DIR := $(VENDOR_DIR)/grpc-gateway
VALIDATE_TMP := $(OPENAPI_TMPL_DIR)/protoc-gen-validate-temp

.PHONY: .bin-deps fetch-protobuf fetch-google-type fetch-openapiv2-options fetch-google-api fetch-validate .vendor-proto

.bin-deps: export GOBIN := $(LOCAL_BIN)
.bin-deps: export PROTOC_VERSION := protoc-31.1-osx-aarch_64
.bin-deps:
	$(info Installing binary dependencies...)

	rm -f $(LOCAL_BIN)/protoc
	rm -rf $(LOCAL_BIN)/include
	rm -f $(PROTOC_VERSION).zip

	curl -LO https://github.com/protocolbuffers/protobuf/releases/download/v31.1/$(PROTOC_VERSION).zip && \
	unzip -o $(PROTOC_VERSION).zip -d $(LOCAL_BIN) && \
	rm $(PROTOC_VERSION).zip

	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && \
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest && \
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest && \
	go install github.com/envoyproxy/protoc-gen-validate@latest

fetch-protobuf:
	rm -rf $(PROTO_TMPL_DIR)
	git clone -b main --single-branch -n --depth=1 --filter=tree:0 \
	https://github.com/protocolbuffers/protobuf $(PROTO_TMPL_DIR)&& \
	cd $(PROTO_TMPL_DIR) && \
	git sparse-checkout set --no-cone src/google/protobuf && \
	git checkout
	rm -rf $(VENDOR_DIR)/google/protobuf && \
	mkdir -p $(VENDOR_DIR)/google
	mv $(PROTO_TMPL_DIR)/src/google/protobuf $(VENDOR_DIR)/google
	rm -rf $(PROTO_TMPL_DIR)

fetch-google-type:
	rm -rf $(TYPE_TMPL_DIR)
	git clone -b master --single-branch -n --depth=1 --filter=tree:0 \
		https://github.com/googleapis/googleapis $(TYPE_TMPL_DIR) && \
	cd $(TYPE_TMPL_DIR) && \
	git sparse-checkout init --cone && \
	git sparse-checkout set google/type && \
	git checkout
	rm -rf $(VENDOR_DIR)/google/type && \
	mkdir -p $(VENDOR_DIR)/google
	mv $(TYPE_TMPL_DIR)/google/type $(VENDOR_DIR)/google
	rm -rf $(TYPE_TMPL_DIR)

fetch-openapiv2-options:
	rm -rf $(OPENAPI_TMPL_DIR)
	git clone -b main --single-branch -n --depth=1 --filter=tree:0 \
	  https://github.com/grpc-ecosystem/grpc-gateway $(OPENAPI_TMPL_DIR) && \
	cd $(OPENAPI_TMPL_DIR) && \
	git sparse-checkout init --cone && \
	git sparse-checkout set protoc-gen-openapiv2/options && \
	git checkout
	rm -rf $(VENDOR_DIR)/protoc-gen-openapiv2
	mkdir -p $(VENDOR_DIR)/protoc-gen-openapiv2
	mv $(OPENAPI_TMPL_DIR)/protoc-gen-openapiv2/options \
	   $(VENDOR_DIR)/protoc-gen-openapiv2/
	rm -rf $(OPENAPI_TMPL_DIR)

fetch-google-api:
	rm -rf $(VENDOR_DIR)/google/api
	git clone -b master --single-branch -n --depth=1 --filter=tree:0 \
	  https://github.com/googleapis/googleapis fetch-googleapi-temp && \
	cd fetch-googleapi-temp && \
	git sparse-checkout init --cone && \
	git sparse-checkout set google/api && \
	git checkout && \
	rm -rf $(VENDOR_DIR)/google/api && \
	mkdir -p $(VENDOR_DIR)/google && \
	mv google/api $(VENDOR_DIR)/google/ && \
	cd .. && rm -rf fetch-googleapi-temp

fetch-validate:
	rm -rf $(VALIDATE_TMP)
	git clone -b v0.6.0 --single-branch -n --depth=1 --filter=tree:0 \
	  https://github.com/envoyproxy/protoc-gen-validate.git $(VALIDATE_TMP) && \
	cd $(VALIDATE_TMP) && \
	git sparse-checkout init --cone && \
	git sparse-checkout set validate && \
	git checkout
	rm -rf $(VENDOR_DIR)/validate
	mkdir -p $(VENDOR_DIR)
	mv $(VALIDATE_TMP)/validate $(VENDOR_DIR)/
	rm -rf $(VALIDATE_TMP)

.vendor-proto: fetch-protobuf fetch-google-type fetch-openapiv2-options fetch-google-api fetch-validate
	@echo "✅ All vendor protos are up to date"

PROTO_PATH := $(CURDIR)/api/proto
API_OUT_PATH := $(CURDIR)/pkg
SWAGGER_OUT_PATH := $(CURDIR)/api/swagger

PROTO_FILES          := $(shell find $(PROTO_PATH)       -type f -name "*.proto"       | sed 's|$(PROTO_PATH)/||')
SERVICE_PROTO_FILES       := $(shell find $(PROTO_PATH) -type f -name "*_service.proto" | sed 's|$(PROTO_PATH)/||')

.PHONY: .protoc-generate
.protoc-generate: .bin-deps .vendor-proto
	mkdir -p $(API_OUT_PATH) $(SWAGGER_OUT_PATH)

	# 1) Go-модели (vendor-proto + proto)
	protoc \
	  -I $(VENDOR_DIR) \
	  -I $(PROTO_PATH) \
	  --plugin=protoc-gen-go=$(LOCAL_BIN)/protoc-gen-go \
	  --go_out=$(API_OUT_PATH) \
	  --go_opt=paths=source_relative \
	  $(PROTO_FILES)

	# 2) gRPC-стабы для сервисов (vendor-proto и proto)
	protoc \
	  -I $(VENDOR_DIR) \
	  -I $(PROTO_PATH) \
	  --plugin=protoc-gen-go-grpc=$(LOCAL_BIN)/protoc-gen-go-grpc \
	  --go-grpc_out=$(API_OUT_PATH) \
	  --go-grpc_opt=paths=source_relative \
	  $(SERVICE_PROTO_FILES)

	# 3) HTTP-gateway
	protoc \
	  -I $(VENDOR_DIR) \
	  -I $(PROTO_PATH) \
	  --plugin=protoc-gen-grpc-gateway=$(LOCAL_BIN)/protoc-gen-grpc-gateway \
	  --grpc-gateway_out=$(API_OUT_PATH) \
	  --grpc-gateway_opt paths=source_relative \
	  --grpc-gateway_opt logtostderr=true \
	  --grpc-gateway_opt generate_unbound_methods=true \
	  $(SERVICE_PROTO_FILES)

	# 4) Swagger/OpenAPI
	protoc \
	  -I $(VENDOR_DIR) \
	  -I $(PROTO_PATH) \
	  --plugin=protoc-gen-openapiv2=$(LOCAL_BIN)/protoc-gen-openapiv2 \
	  --openapiv2_out=$(SWAGGER_OUT_PATH) \
	  --openapiv2_opt logtostderr=true \
	  --openapiv2_opt allow_merge=true \
	  --openapiv2_opt generate_unbound_methods=true \
	  $(SERVICE_PROTO_FILES)

	protoc \
		-I $(VENDOR_DIR) \
		-I $(PROTO_PATH) \
		--plugin=protoc-gen-validate=$(LOCAL_BIN)/protoc-gen-validate \
		--validate_out="lang=go,paths=source_relative:$(API_OUT_PATH)" \
		$(PROTO_FILES) $(SERVICE_PROTO_FILES)

	go mod tidy


gen:
	mkdir -p $(API_OUT_PATH) $(SWAGGER_OUT_PATH)

	# 1) Go-модели (vendor-proto + proto)
	protoc \
	  -I $(VENDOR_DIR) \
	  -I $(PROTO_PATH) \
	  --plugin=protoc-gen-go=$(LOCAL_BIN)/protoc-gen-go \
	  --go_out=$(API_OUT_PATH) \
	  --go_opt=paths=source_relative \
	  $(PROTO_FILES)

	# 2) gRPC-стабы для сервисов (vendor-proto и proto)
	protoc \
	  -I $(VENDOR_DIR) \
	  -I $(PROTO_PATH) \
	  --plugin=protoc-gen-go-grpc=$(LOCAL_BIN)/protoc-gen-go-grpc \
	  --go-grpc_out=$(API_OUT_PATH) \
	  --go-grpc_opt=paths=source_relative \
	  $(SERVICE_PROTO_FILES)

	# 3) HTTP-gateway
	protoc \
	  -I $(VENDOR_DIR) \
	  -I $(PROTO_PATH) \
	  --plugin=protoc-gen-grpc-gateway=$(LOCAL_BIN)/protoc-gen-grpc-gateway \
	  --grpc-gateway_out=$(API_OUT_PATH) \
	  --grpc-gateway_opt paths=source_relative \
	  --grpc-gateway_opt logtostderr=true \
	  --grpc-gateway_opt generate_unbound_methods=true \
	  $(SERVICE_PROTO_FILES)

	# 4) Swagger/OpenAPI
	protoc \
	  -I $(VENDOR_DIR) \
	  -I $(PROTO_PATH) \
	  --plugin=protoc-gen-openapiv2=$(LOCAL_BIN)/protoc-gen-openapiv2 \
	  --openapiv2_out=$(SWAGGER_OUT_PATH) \
	  --openapiv2_opt logtostderr=true \
	  --openapiv2_opt allow_merge=true \
	  --openapiv2_opt generate_unbound_methods=true \
	  $(SERVICE_PROTO_FILES)

	protoc \
		-I $(VENDOR_DIR) \
		-I $(PROTO_PATH) \
		--plugin=protoc-gen-validate=$(LOCAL_BIN)/protoc-gen-validate \
		--validate_out="lang=go,paths=source_relative:$(API_OUT_PATH)" \
		$(PROTO_FILES) $(SERVICE_PROTO_FILES)

	go mod tidy


include .env

DB_USER      = $(PG_SUPER_USER)
DB_PASSWORD  = $(PG_SUPER_PASSWORD)
DB_HOST      = $(PG_HOST)
DB_PORT      = $(PG_MASTER_PORT)
DB_NAME      = $(PG_DATABASE)

DATABASE_DSN := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

GOOSE_BIN := $(shell go env GOPATH)/bin/goose

.PHONY: install-goose
install-goose:
	go install github.com/pressly/goose/v3/cmd/goose@latest

.PHONY: up down reset status create

up:
	$(GOOSE_BIN) -dir migrations postgres "$(DATABASE_DSN)" up

down:
	$(GOOSE_BIN) -dir migrations postgres "$(DATABASE_DSN)" down

reset:
	$(GOOSE_BIN) -dir migrations postgres "$(DATABASE_DSN)" reset

status:
	$(GOOSE_BIN) -dir migrations postgres "$(DATABASE_DSN)" status

create:
	$(GOOSE_BIN) -dir migrations create $(name) sql

.PHONY: docker-up start docker-down

docker-up:
	docker compose up -d

docker-down:
	docker compose down -v

wait-db:
	@echo "Waiting for master on localhost:5432…"
	@until pg_isready -h localhost -p 5432 -U gopher >/dev/null 2>&1; do \
	    printf "."; sleep 1; \
	done
	@echo "\nWaiting for replica1 on localhost:5433…"
	@until pg_isready -h localhost -p 5433 -U gopher >/dev/null 2>&1; do \
	    printf "."; sleep 1; \
	done
	@echo "\nWaiting for replica2 on localhost:5434…"
	@until pg_isready -h localhost -p 5434 -U gopher >/dev/null 2>&1; do \
	    printf "."; sleep 1; \
	done
	@echo "\nAll databases are ready!"

start-app: docker-up wait-db up
	go run ./cmd/pvz/main.go ./cmd/pvz/setup_flag.go --config ./configs/config.yaml --env .env

notifier:
	go run ./internal/services/notifier/cmd/notifier/main.go


# Пример команды:
# make mocks INTERFACE=./internal/usecase.OrdersRepository OUT=./internal/usecase/mock/orders_repo_mock.go MOCK_NAME=OrdersRepositoryMock
mocks:
	@echo "=> Generating mock for $(INTERFACE) into $(OUT)"
	@mkdir -p $(dir $(OUT))
	minimock -i $(INTERFACE) -o $(OUT) -n $(MOCK_NAME)




.PHONY: tests coverage

tests:
	go test -tags "integration e2e" -count=5 ./...


unit-tests:
	go test -v -count=5 ./...


integration-tests:
	go test -tags=integration ./...


e2e-tests:
	go test -tags=e2e ./...

# Путь для сохранения отчётов покрытия
COVDIR ?= coverage
COVFILE := $(COVDIR)/coverage.out
COVHTML := $(COVDIR)/coverage.html

coverage:
	@mkdir -p $(COVDIR)
	go test -v -coverprofile=$(COVFILE) ./...
	go tool cover -html=$(COVFILE) -o $(COVHTML)

local-run:
	go run ./cmd/pvz/main.go ./cmd/pvz/setup_flag.go --config ./configs/config.yaml --env .env

