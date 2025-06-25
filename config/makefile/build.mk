.PHONY: build\:local build\:release

build\:local:
	docker compose --profile local up --build -d

build\:release:
	git tag v$(V)
	@read -p "Press enter to confirm and push to origin ..." && git push origin v$(V)
