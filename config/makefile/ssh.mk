___SSH___ROOT___PATH      := $(shell pwd)
___SSH___SSH___DEV___FILE := ___SSH___ROOT___PATH/.ssh/dev/key
___SSH___SSH___PRD___FILE := ___SSH___ROOT___PATH/.ssh/prd/key

.ssh\:dev:
	echo email
	#ssh-keygen -t rsa -b 4096 -C (email) -f ___SSH___SSH___DEV___FILE

.key\:prd:
	ssh-keygen -t rsa -b 4096 -C (email) -f ___SSH___SSH___PRD___FILE
