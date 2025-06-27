.PHONY: fresh audit watch format

format:
	gofmt -w -s .

fresh:
	docker compose down --remove-orphans && \
	docker container prune -f && \
	docker image prune -f && \
	docker volume prune -f && \
	docker network prune -f && \
	docker system prune -a --volumes -f && \
	docker ps -aq | xargs --no-run-if-empty docker stop && \
	docker ps -aq | xargs --no-run-if-empty docker rm && \
	docker ps

audit:
	$(call external_deps,'.')
	$(call external_deps,'./app/...')
	$(call external_deps,'./database/...')
	$(call external_deps,'./docs/...')

watch:
	# --- Works with (air).
	#     https://github.com/air-verse/air
	cd $(APP_PATH) && air

install-air:
	# --- Works with (air).
	#     https://github.com/air-verse/air
	@echo "Installing air ..."
	@go install github.com/air-verse/air@latest
