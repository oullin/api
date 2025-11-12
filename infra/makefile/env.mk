.PHONY: env\:check env\:fresh env\:init env\:print

___ENV___ROOT__PATH        := $(shell pwd)
___ENV___FILE_NAME         := ".env"
___ENV___EXAMPLE_FILE_NAME := ".env.example"

-include $(___ENV___ROOT__PATH)/.env

env\:init:
	@if [ ! -f $(___ENV___FILE_NAME) ]; then \
	  printf "The $(BLUE)[.env]$(NC) file not found in the current directory.\n"; \
	  printf "New $(GREEN)[.env]$(NC) file successfully create from the $(YELLOW)[.env.example]$(NC) file using the $(BLUE)[make env:init]$(NC).\n"; \
	  cp $(___ENV___EXAMPLE_FILE_NAME) $(___ENV___FILE_NAME) ; \
	fi

env\:check:
	@if [ ! -f $(___ENV___FILE_NAME) ]; then \
	  printf "$(RED)Error:$(NC) $(BLUE)[.env]$(NC) file not found in the current directory. Please run $(BLUE)[env:init]$(NC) first.\n"; \
	else \
	  printf "$(BLUE)[.env]$(NC) file found.\n"; \
	fi

env\:fresh:
	@if [ -f $(___ENV___FILE_NAME) ]; then \
    	printf "\nRemoving development environment file $(YELLOW)[$(___ENV___FILE_NAME)]$(NC).\n"; \
    	rm -f $(___ENV___FILE_NAME); \
    	printf "Finished removing $(YELLOW)[$(___ENV___FILE_NAME)]$(NC) ($(RED)if it existed$(NC)).\n\n----\n"; \
    	$(MAKE) env:init; \
    else \
    	$(MAKE) env:init; \
    fi

env\:print:
	@echo "APP NAME ........ : $(ENV_APP_NAME)		"
	@echo "APP TYPE ........ : $(ENV_APP_ENV_TYPE)	"
	@echo "ROOT_NETWORK .... : $(ROOT_NETWORK)      "
	@echo " --------------------------------------- "
	@echo "DB_MIGRATE_PATH . : $(DB_MIGRATE_PATH)	"
	@echo "DB_MIGRATE_VOL_MAP: $(DB_MIGRATE_VOL_MAP)"
