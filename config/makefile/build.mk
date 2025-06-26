.PHONY: build\:local build\:prod build\:release

build\:local:
	docker compose --profile local up --build -d

build\:prod:
	docker compose --profile prod up --build -d

build\:release:
	@printf "\n$(YELLOW)Tagging images to be released.$(NC)\n"
	docker tag api-api ghcr.io/gocanto/oullin_api:0.0.1 && \
	docker tag api-caddy_prod ghcr.io/gocanto/oullin_proxy:0.0.1

	@printf "\n$(CYAN)Pushing release to GitHub registry.$(NC)\n"
	docker push ghcr.io/gocanto/oullin_api:0.0.1 && \
	docker push ghcr.io/gocanto/oullin_proxy:0.0.1
