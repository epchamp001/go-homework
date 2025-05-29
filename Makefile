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
	go build -o $(BINARY_NAME) cmd/pvz/main.go

start:
	./$(BINARY_NAME)

# Обновить зависимости, запустить линтеры, собрать и запустить приложение
run: update linter build start
