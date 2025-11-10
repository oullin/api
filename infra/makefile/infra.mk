.PHONY: supv\:api\:status supv\:api\:start supv\:api\:stop
.PHONY: supv\:api\:stop supv\:api\:restart supv\:api\:logs supv\:api\:logs-err
.PHONY: apparmor-complain apparmor-enforce

___API__SUPERVISOR := oullin-api

supv\:api\:status:
	@sudo supervisorctl status $(___API__SUPERVISOR)

supv\:api\:start:
	@sudo supervisorctl start $(___API__SUPERVISOR)

supv\:api\:stop:
	@sudo supervisorctl stop $(___API__SUPERVISOR)

supv\:api\:restart:
	@sudo supervisorctl restart $(___API__SUPERVISOR)

supv\:api\:logs:
	@sudo tail -f /var/log/supervisor/$(___API__SUPERVISOR).log

supv\:api\:logs-err:
	@sudo tail -f /var/log/supervisor/$(___API__SUPERVISOR).err.log


apparmor-complain:
	@sudo aa-complain /etc/apparmor.d/docker-default

apparmor-enforce:
	@sudo aa-enforce /etc/apparmor.d/docker-default
