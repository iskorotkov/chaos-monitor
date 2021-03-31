VERSION = v0.6.0
IMAGE = iskorotkov/chaos-monitor

.PHONY: ci
ci: build test-short test build-image push-image

.PHONY: build
build:
	go build ./...

.PHONY: test
test:
	go test -v ./...

.PHONY: test-short
test-short:
	go test -short -v ./...

.PHONY: build-image
build-image:
	docker build -f build/monitor.dockerfile -t $(IMAGE):$(VERSION) .

.PHONY: push-image
push-image:
	docker push $(IMAGE):$(VERSION)
