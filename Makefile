PROJECT_ROOT?=$(shell pwd)
PROJECT_PKG?=evhub
TARGET_PKG=$(PROJECT_PKG)/cmd
IMAGE_PREFIX?=evhub
TARGET_IMAGE=$(IMAGE_PREFIX)/$(TARGET):$(VERSION)

lint:
	build/lint.sh

fmt:
	gofmt -l -s -w .
	goimports -w .

proto:
	build/generate-protobuf.sh

build:
	make proto
	go build -o ./bin/$(TARGET) ./cmd/$(TARGET)

binary:
	CGO_ENABLED=0 go build  \
	 -o /usr/local/go/src/$(PROJECT_PKG)/bin/$(TARGET) /usr/local/go/src/$(TARGET_PKG)/$(TARGET)

target:
	mkdir -p $(PROJECT_ROOT)/bin
	docker run --rm -i -v $(PROJECT_ROOT):/usr/local/go/src/$(PROJECT_PKG) \
	  -w /usr/local/go/src/$(PROJECT_PKG) golang:1.17.12 \
    make binary TARGET=$(TARGET)

image:
	make target TARGET=$(TARGET) && \
	temp=`mktemp -d` && \
	cp -r $(PROJECT_ROOT)/build/images/$(TARGET)/ $$temp/$(TARGET)/  && cp -r $(PROJECT_ROOT)/bin/ $$temp/$(TARGET)/bin/ && \
	docker build -t $(TARGET_IMAGE) $$temp/$(TARGET) && \
	rm -r $$temp
	rm -r ./bin

start:
	build/docker-compose/start.sh

stop:
	build/docker-compose/stop.sh

.PHONY: image target clean  binary
