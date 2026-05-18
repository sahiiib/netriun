APP=netriun

build:
	go build -o $(APP) ./cmd/server

run:
	go run ./cmd/server

docker:
	docker build -t netriun/netriun-web:v1 .

push:
	docker push netriun/netriun-web:v1

k8s:
	kubectl apply -f deployments/k8s/

clean:
	rm -f $(APP)