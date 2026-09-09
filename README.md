# Netriun

AI-Native Networking Infrastructure Platform

## Run

```bash
go run ./cmd/server
```

## Container

```bash
make docker
make push
```

The default image is `docker.io/sahiiib/netriun-web:latest`.

## Kubernetes and Cloudflare credentials

The real `deployments/k8s/cloudflare-secret.yaml` and `.env` files are local,
gitignored files. Never commit credentials or put them under `web/static`,
which is publicly served. Deployment manifests and environment files are
excluded from the Docker build context.

For a new installation:

1. Apply `deployments/k8s/namespace.yaml`.
2. Copy `deployments/k8s/cloudflare-secret.example.yaml` to
   `deployments/k8s/cloudflare-secret.yaml` and replace the placeholder locally.
3. Run `chmod 600 deployments/k8s/cloudflare-secret.yaml`.
4. Apply that local Secret with
   `kubectl apply -f deployments/k8s/cloudflare-secret.yaml`.
5. Run `make k8s`.

`make k8s` deploys the application, services and Cloudflare connector. It does
not apply the example Secret or overwrite existing credentials. If using nginx
Ingress, apply `deployments/k8s/ingress.yaml` separately.

## Previously committed token

Removing a file from Git tracking does not remove it from existing commits or
GitHub history. Rotate the exposed tunnel token in Cloudflare, update the local
Secret, apply it, and restart the connector with
`kubectl rollout restart deployment/cloudflared -n netriun`.
Follow Cloudflare's compromised-token procedure to disconnect existing
connections as well:
https://developers.cloudflare.com/tunnel/advanced/tunnel-tokens/

Removing the old value from GitHub history additionally requires coordinated
history rewriting and a force-push; it is separate from these working-tree
changes. Other clones and cached copies may retain the original commits.
