___SSH___ROOT___PATH      := $(shell pwd)
___SSH___SSH___DEV___FILE := $(___SSH___ROOT___PATH)/.ssh/dev/key
___SSH___SSH___PRD___FILE := $(___SSH___ROOT___PATH)/.ssh/prd/key

ssh\:dev:
	rm -rf ___SSH___SSH___DEV___FILE && \
	rm -rf "$(___SSH___SSH___DEV___FILE).pub" && \
	ssh-keygen -t rsa -b 4096 -C $(email) -f $(___SSH___SSH___DEV___FILE)

ssh\:prd:
	rm -rf ___SSH___SSH___PRD___FILE && \
	rm -rf "$(___SSH___SSH___PRD___FILE).pub" && \
	ssh-keygen -t rsa -b 4096 -C $(email) -f $(___SSH___SSH___PRD___FILE)
