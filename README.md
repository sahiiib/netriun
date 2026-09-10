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
5. Complete the content-editor first installation below, then run `make k8s`.

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

## Content editor

The editor is at `https://netriun.com/neditport2065` (username `admin`).
A hidden URL is not the authentication mechanism: the editor uses bcrypt
password verification, database-backed expiring sessions, secure HttpOnly
cookies, origin and CSRF checks, and shared login rate limits.

- Products: create, edit, reorder, remove, hide, mark coming soon or publish.
- News: draft/publish articles with text, date and image URL, shown at `/news`.
- About: edit introduction, brand ownership, mission and approach.
- Partners: names, descriptions, image URLs, links and draft/published state.
- Contact settings: separate Sales, Support and Info recipients.
- Inbox: latest 100 contact messages and email delivery status.

Content has English, German, Russian and Armenian fields. English is required;
blank translations fall back to English. Images use HTTPS URLs or existing
`/static/img/` assets; direct file uploads and rich HTML are not supported.
Published edits take effect immediately without rebuilding or redeploying.
`Save changes` saves every section; drafts remain hidden. Concurrent saves
from stale editor sessions are rejected so one editor cannot silently overwrite
another. Changes made in the editor live in PostgreSQL, not in Git.

### First installation

```bash
python3 scripts/provision-editor.py
kubectl apply -f deployments/k8s/editor-secret.yaml
kubectl apply -f deployments/k8s/editor-database.yaml -f deployments/k8s/editor-backup.yaml
kubectl rollout status deployment/netriun-editor-postgres -n netriun
make k8s
```

The provisioning script refuses to overwrite existing credentials. The initial
password is saved only in `~/.config/netriun/editor-credentials.txt` (mode 0600).
The Kubernetes Secret manifest is gitignored. Do not paste either into issues
or commits. To reset access, generate a new bcrypt hash with `go run
./cmd/adminhash` using stdin, update `ADMIN_PASSWORD_HASH` in the local Secret,
apply it, revoke sessions with `DELETE FROM admin_sessions` in PostgreSQL,
and restart `deployment/netriun-web`. Update the private credential file too.

The database is independent from Nexus and uses a Retain PV on `k8s-node01`
at `/var/lib/netriun-web/postgresql`. Both web replicas share it. Readiness
checks include database connectivity; liveness checks do not restart the app
just because the database is briefly unavailable. If the database node is down,
content pages will be unavailable until it recovers or the data is restored.

The daily backup runs at 02:15 Asia/Yerevan and retains 14 days of custom-format
`pg_dump` files under `/var/lib/netriun-web/backups` on that same node. These
backups protect against accidental edits but not loss of the node/disk; copy
them to independent storage for disaster recovery. Restore to an empty database
with `pg_restore --no-owner --no-acl -U netriun -d netriun <backup.dump>` while
the web deployment is stopped. The backup includes contact messages; restrict
access to the files. Do not commit database dumps.

### Contact email

All three default recipients are `s.amrei@netriun.com`; edit them in the panel.
Messages are always stored in the inbox. Without SMTP they are marked
`received`, not claimed to have been emailed. To enable mail delivery, add
`SMTP_HOST`, `SMTP_PORT`, `SMTP_USERNAME`, `SMTP_PASSWORD` and `SMTP_FROM` to
the local editor Secret, apply it, and restart the web deployment. Port 587
uses required STARTTLS; port 465 uses implicit TLS. A trusted TLS certificate
is required. The visitor's address goes in Reply-To, never From.

`emailed` means the SMTP server accepted the message; `email-failed` means it
is still available in the inbox but delivery failed. Existing messages are not
automatically resent when SMTP is configured. There is currently no background
retry queue. Credentials are never editable through the website.

### Tests

```bash
go test ./...
go vet ./...
# Use a disposable PostgreSQL database only:
TEST_DATABASE_URL='postgres://user:password@localhost:5432/test?sslmode=disable' go test -race ./internal/cms
# Run a local app with a disposable DB and test password, then:
TEST_ADMIN_PASSWORD='your-test-password' python3 scripts/test-editor.py
```

The integration test resets its test tables; never point it at production.
The browser/end-to-end test script refuses non-local origins.
