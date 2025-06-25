.PHONY: build\:app build\:release

build\:app:
	docker compose up --build -d caddy

build\:release:
	git tag v$(V)
	@read -p "Press enter to confirm and push to origin ..." && git push origin v$(V)
