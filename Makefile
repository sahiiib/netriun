APP=netriun
IMAGE=docker.io/sahiiib/netriun-web:latest

build:
	go build -o $(APP) ./cmd/server

run:
	go run ./cmd/server

docker:
	docker build -t $(IMAGE) .

push:
	docker push $(IMAGE)

k8s:
	kubectl apply -f deployments/k8s/

clean:
	rm -f $(APP)
