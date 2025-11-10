
define external_deps
	@echo '-- $(1)';  go list -f '{{join .Deps "\n"}}' $(1) | grep -v github.com/$(REPO_OWNER)/blog | xargs go list -f '{{if not .Standard}}{{.ImportPath}}{{end}}'
endef

# --- Padding left or Right
define padding_lr
	@printf "%*s" $(1) ""
endef
