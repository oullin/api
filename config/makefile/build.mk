.PHONY: build\:app build\:flush build\:app\:linux build\:release build\:run build\:fresh

___BIN___ROOT__PATH           := $(shell pwd)
___BIN___ENV___FILE__TEMPLATE := .env.production
___BIN___FULL__PATH           := $(___BIN___ROOT__PATH)/bin
___BIN___ENV___FILE           := $(___BIN___FULL__PATH)/.env
___BIN___APP___FILE           := $(___BIN___FULL__PATH)/app

# --- Storage
___BIN___STORAGE__PATH := $(___BIN___FULL__PATH)/storage
___BIN___LOGS__PATH    := $(___BIN___STORAGE__PATH)/logs


build\:fresh:
	make build:app && make build:run

build\:run:
	cd $(___BIN___FULL__PATH) && ./app

#
build\:app\:linux:
	@printf "\n$(BLUE)[BIN]$(NC) Building the app in [amd64] has started.\n"
	make build:env && \
	make build:flush && \
	cd $(APP_PATH) && \
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o "$(ROOT_PATH)/bin/app_linux" -ldflags='-X main.Version=$(VERSION) -extldflags "-static"' -tags '$(DATABASE) $(SOURCE)' $(APP_PATH)

build\:release:
	git tag v$(V)
	@read -p "Press enter to confirm and push to origin ..." && git push origin v$(V)

build\:app:
	@printf "\n$(BLUE)[BIN]$(NC) Building the app has started.\n"
	make build:env && \
	make build:flush && \
	CGO_ENABLED=0 go build -a -ldflags='-X main.Version=$(VERSION)' -o "$(ROOT_PATH)/bin/app" -tags '$(DATABASE) $(SOURCE)' $(APP_PATH)
	@printf "$(GREEN)[BIN]$(NC) Building the app has finished.\n\n"

build\:flush:
	@printf "$(BLUE)[BIN]$(NC) Flushing the previous builds has started.\n"
	@sleep 1
	rm -f $(___BIN___ENV___FILE)
	rm -f $(___BIN___APP___FILE)
	rm -rf $(___BIN___LOGS__PATH)
	mkdir -m 755 $(___BIN___LOGS__PATH)
	touch $(___BIN___LOGS__PATH)/.gitkeep
	@printf "$(GREEN)[BIN]$(NC) Flushing has finished.\n\n"

build\:env:
	cp $(___BIN___ENV___FILE__TEMPLATE) $(___BIN___ENV___FILE)
