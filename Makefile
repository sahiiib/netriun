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
	kubectl apply -f deployments/k8s/namespace.yaml
	kubectl apply -f deployments/k8s/editor-database.yaml -f deployments/k8s/editor-backup.yaml
	kubectl apply -f deployments/k8s/deployment.yaml -f deployments/k8s/service-clusterip.yaml -f deployments/k8s/service-nodeport.yaml -f deployments/k8s/cloudflare-deployment.yaml

clean:
	rm -f $(APP)
