.PHONY: supv\:api\:status supv\:api\:start supv\:api\:stop
.PHONY: supv\:api\:stop supv\:api\:restart supv\:api\:logs supv\:api\:logs-err

___API__SUPERVISOR := oullin-api

supv\:api\:status:
	@printf "\n$(YELLOW)[supervisor]$(NC) - $(CYAN)$(___API__SUPERVISOR) status.$(NC)"
	@sudo supervisorctl status $(___API__SUPERVISOR)

supv\:api\:start:
	@printf "\n$(YELLOW)[supervisor]$(NC) - $(CYAN)$(___API__SUPERVISOR) start command sent.$(NC)"
	@sudo supervisorctl start $(___API__SUPERVISOR)

supv\:api\:stop:
	@printf "\n$(YELLOW)[supervisor]$(NC) - $(CYAN)$(___API__SUPERVISOR) stop command sent.$(NC)"
	@sudo supervisorctl stop $(___API__SUPERVISOR)

supv\:api\:restart:
	@printf "\n$(YELLOW)[supervisor]$(NC) - $(CYAN)$(___API__SUPERVISOR) restart command sent.$(NC)"
	@sudo supervisorctl restart $(___API__SUPERVISOR)

supv\:api\:logs:
	@printf "\n$(YELLOW)[supervisor]$(NC) - $(CYAN)$(___API__SUPERVISOR) logs. (Press Ctrl+C to exit)$(NC)"
	@sudo tail -f /var/log/supervisor/$(___API__SUPERVISOR).log

supv\:api\:logs-err:
	@printf "\n$(YELLOW)[supervisor]$(NC) - $(CYAN)$(___API__SUPERVISOR) error logs. (Press Ctrl+C to exit)$(NC)"
	@sudo tail -f /var/log/supervisor/$(___API__SUPERVISOR).err.log
