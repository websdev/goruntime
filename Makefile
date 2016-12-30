.PHONY: bootstrap
bootstrap:
	script/install-glide
	glide install

.PHONY: update
update:
	glide up
	glide install

.PHONY: compile-test
compile-test:
	go test -v -cover -race $(shell glide nv)

.PHONY: docker-test
docker-test:
	docker run goruntime make compile-test
