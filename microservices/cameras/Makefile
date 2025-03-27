PROJECT=$(shell cat go.mod | grep module | awk -F ' ' '{print $$2}' | awk -F '/' '{print $$NF}')

AUTOGEN=autogen
DOCS=$(AUTOGEN)/docs
TEST=$(AUTOGEN)/test
REPORTS=.reports
VENDOR=vendor

help: ## Показать описание основных команд
	@echo "\033[1;34m$(PROJECT) usage:\033[0m"
	@egrep -h '\s##\s' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

setup: dependencies ## Настроить рабочее окружение
	@echo -n "\033[1;34mSetup...\033[0m\t"
	@mkdir -p $(AUTOGEN)/server
	@mkdir -p $(AUTOGEN)/docs
	@mkdir -p $(TEST)
	@mkdir -p $(REPORTS)
	@echo "\033[0;32mOK\033[0m"

clean: ## Очистить рабочее окружение
	@echo -n "\033[1;34mClean...\033[0m\t"
	@golangci-lint cache clean
	@rm -rf $(VENDOR)
	@rm -rf $(REPORTS)
	@rm -rf $(TEST)
	@rm -rf $(DOCS)
	@rm -rf $(AUTOGEN)
	@go clean -r -i -testcache -modcache
	@echo "\033[0;32mOK\033[0m"

update: setup ## Обновить зависимости
	@echo -n "\033[1;34mUpdate...\033[0m\t"
	@go mod tidy
	@go mod vendor
	@echo "\033[0;32mOK\033[0m"

format: setup ## Запуск форматирования исходного кода
	@echo -n "\033[1;34mFormat...\033[0m\t"
	@golangci-lint run --fix ./internal/...
	@echo "\033[0;32mOK\033[0m"

lint: format ## Запуск линтера
	@echo -n "\033[1;34mLinter...\033[0m\t"
	@golangci-lint run ./internal/... --out-format checkstyle > $(REPORTS)/golangci-lint.xml
	@echo "\033[0;32mOK\033[0m"

mock-gen: setup ## Генерация mock
	@echo -n "\033[1;34mMocks...\033[0m\t"
	@go generate ./...
	@echo "\033[0;32mOK\033[0m"

test: mock-gen ## Запуск тестов
	@echo -n "\033[1;34mTest...\033[0m\n"
	@go test -v ./internal/...
	@echo "\033[0;32mOK\033[0m"

test-cover: mock-gen # Запуск покрытия тестами
	@echo -n "\033[1;34mTest coverage...\033[0m\n"
	@go test -coverprofile=$(TEST)/coverage.out ./internal/...
	@go tool cover -html=$(TEST)/coverage.out -o $(TEST)/coverage.html
	@echo "Coverage report generated in $(TEST)/coverage.html"
	@echo "\033[0;32mOK\033[0m"

swagger: swagger-code-gen ## Генерация сваггера

swagger-code-gen: swagger-spec-gen
	@echo -n "\033[1;34mCodegen...\033[0m\t"
	@oapi-codegen --config=scripts/codegen.yml $(DOCS)/swagger.yaml
	@echo "\033[0;32mOK\033[0m"

swagger-spec-gen: setup
	@echo -n "\033[1;34mSpec...\033[0m\t\t"
	@sh ./scripts/spec-gen.sh $(DOCS) cmd/$(PROJECT),internal/service
	@sh ./scripts/converter.sh $(DOCS)/swagger.yaml
	@echo "\033[0;32mOK\033[0m"

dependencies: ## Установить зависимости
	@echo -n "\033[1;34mDependencies...\033[0m\t"
	@command -v oapi-codegen > /dev/null || go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
	@echo "\033[0;32mOK\033[0m"

all: swagger update mock-gen test test-cover ## Последовательный запуск основных команд

.DEFAULT_GOAL=all
