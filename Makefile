-include .env
export

DOCKER ?= true

RUN = docker run --rm -it \
		-v $(PWD):/usr/src/app \
		-v $(PWD)/runtime:/usr/src/app/runtime \
		--env-file .env \
		$(PROJECT)

ifeq ($(DOCKER), false)
	RUN = go
endif

ARGS ?=

## Docker build.
.PHONY: d-build
d-build:
	docker build -t $(PROJECT) .

.PHONY: test
test:
	$(RUN) test ./...

.PHONY: run
run:
	$(RUN) $(ARGS)

%:
	@:

