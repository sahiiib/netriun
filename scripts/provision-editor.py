#!/usr/bin/env python3
"""Generate credentials locally without printing passwords or changing the cluster."""
import json
import os
from pathlib import Path
import secrets
import subprocess

os.umask(0o077)
root = Path(__file__).resolve().parents[1]
secret_file = root / 'deployments/k8s/editor-secret.yaml'
credentials = Path.home() / '.config/netriun/editor-credentials.txt'
if secret_file.exists() or credentials.exists():
    raise SystemExit('Credentials already exist; refusing to overwrite them.')
password = secrets.token_urlsafe(30)
db_password = secrets.token_urlsafe(36)
password_hash = subprocess.check_output(['go', 'run', './cmd/adminhash'], cwd=root, input=password.encode()).decode()
secret = {'apiVersion':'v1', 'kind':'Secret', 'metadata':{'name':'netriun-editor-secret','namespace':'netriun'}, 'type':'Opaque', 'stringData':{'POSTGRES_PASSWORD':db_password, 'DATABASE_URL':f'postgres://netriun:{db_password}@netriun-editor-postgres:5432/netriun?sslmode=disable', 'ADMIN_PASSWORD_HASH':password_hash}}
secret_file.write_text(json.dumps(secret, indent=2)+'\n')
credentials.parent.mkdir(parents=True, exist_ok=True)
credentials.write_text(f'URL: https://netriun.com/neditport2065\nUsername: admin\nPassword: {password}\n')
print(f'Local secret created: {secret_file}')
print(f'Login credentials saved privately: {credentials}')
