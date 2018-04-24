IMAGE ?= lilic/kube-start-stop
TAG ?= $(shell git describe --tags --always)

build:
		go build -i github.com/lilic/kube-start-stop

linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build github.com/lilic/kube-start-stop

image: linux
	docker build -t "$(IMAGE):$(TAG)" .

deploy:
	kubectl apply -f manifests/

create:
	kubectl apply -f example.yml

undo:
	kubectl delete -f example.yml

delete:
	kubectl delete -f manifests/

test:
	go test -v $(shell go list ./... | grep -v /vendor/ | grep -v /test/)


.PHONY: build linux image deploy delete test
