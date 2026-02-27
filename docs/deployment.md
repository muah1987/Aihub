# Deployment Guide

## Docker Compose (Recommended)

The simplest way to deploy Aihub with all dependencies.

### 1. Prepare environment

```bash
git clone https://github.com/muah1987/Aihub.git
cd Aihub
cp .env.example .env
```

Edit `.env` with production values:

```env
# CRITICAL: Change these for production
JWT_SECRET=<random-64-char-string>
ENCRYPTION_KEY=<random-32-byte-hex-string>
ENVIRONMENT=production

# Database
DATABASE_URL=postgres://aihub:STRONG_PASSWORD@postgres:5432/aihub?sslmode=disable

# CORS — set to your actual domain
CORS_ALLOWED_ORIGINS=https://yourdomain.com

# SMTP (for email verification)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM=noreply@yourdomain.com
APP_BASE_URL=https://yourdomain.com
```

### 2. Generate secure secrets

```bash
# JWT secret (64 random characters)
openssl rand -base64 48

# Encryption key (32 bytes = 64 hex characters)
openssl rand -hex 32
```

### 3. Build and start

```bash
docker compose up -d --build
```

Services:
- **postgres**: PostgreSQL 16 on port 5432
- **server**: Go API server on port 8080
- **web**: React app served via nginx on port 5173

### 4. Verify

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

---

## Manual Deployment (VPS)

### Prerequisites

- Go 1.24+
- Node.js 20+
- PostgreSQL 16
- nginx (reverse proxy)
- Docker (optional, for terminal sandbox)

### 1. Build the backend

```bash
cd Aihub
go build -o bin/aihub-server ./cmd/server
```

### 2. Build the frontend

```bash
cd web
npm ci
npm run build
# Output in web/dist/
```

### 3. Set up PostgreSQL

```bash
sudo -u postgres createuser aihub
sudo -u postgres createdb aihub -O aihub
sudo -u postgres psql -c "ALTER USER aihub PASSWORD 'your-password';"
```

### 4. Configure environment

Export environment variables or create a `.env` file alongside the binary:

```bash
export DATABASE_URL="postgres://aihub:your-password@localhost:5432/aihub?sslmode=disable"
export JWT_SECRET="your-jwt-secret"
export ENCRYPTION_KEY="your-32-byte-hex-key"
export ENVIRONMENT="production"
export CORS_ALLOWED_ORIGINS="https://yourdomain.com"
```

### 5. Run the server

```bash
./bin/aihub-server
```

Migrations run automatically on startup.

### 6. Systemd service (recommended)

Create `/etc/systemd/system/aihub.service`:

```ini
[Unit]
Description=Aihub API Server
After=network.target postgresql.service

[Service]
Type=simple
User=aihub
WorkingDirectory=/opt/aihub
ExecStart=/opt/aihub/bin/aihub-server
EnvironmentFile=/opt/aihub/.env
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now aihub
```

### 7. Nginx reverse proxy

```nginx
server {
    listen 80;
    server_name yourdomain.com;

    # Frontend (static files)
    location / {
        root /opt/aihub/web/dist;
        try_files $uri $uri/ /index.html;
    }

    # API proxy
    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # WebSocket proxy
    location /api/v1/projects/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_read_timeout 86400;
    }

    location /api/v1/notifications/ws {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_read_timeout 86400;
    }

    # GitHub webhook ingress
    location /webhooks/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### 8. TLS with Let's Encrypt

```bash
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d yourdomain.com
```

---

## Environment Variables Reference

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `SERVER_PORT` | No | `8080` | HTTP server port |
| `SERVER_HOST` | No | `0.0.0.0` | Bind address |
| `ENVIRONMENT` | No | `development` | `development` or `production` |
| `DATABASE_URL` | Yes | — | PostgreSQL connection string |
| `JWT_SECRET` | Yes | — | Secret for signing JWT tokens |
| `JWT_ACCESS_TOKEN_TTL` | No | `15m` | Access token lifetime |
| `JWT_REFRESH_TOKEN_TTL` | No | `7d` | Refresh token lifetime |
| `ENCRYPTION_KEY` | Yes | — | 32-byte hex string for AES-256-GCM |
| `CORS_ALLOWED_ORIGINS` | No | `http://localhost:5173` | Comma-separated allowed origins |
| `SMTP_HOST` | No | — | SMTP server hostname |
| `SMTP_PORT` | No | `587` | SMTP server port |
| `SMTP_USERNAME` | No | — | SMTP auth username |
| `SMTP_PASSWORD` | No | — | SMTP auth password |
| `SMTP_FROM` | No | `noreply@aihub.dev` | From address for emails |
| `APP_BASE_URL` | No | `http://localhost:5173` | Base URL for email links |
| `DOCKER_HOST` | No | `unix:///var/run/docker.sock` | Docker daemon socket |
| `SANDBOX_IMAGE` | No | `aihub-sandbox:latest` | Docker image for terminal sandbox |
| `SANDBOX_MEMORY_LIMIT` | No | `256m` | Memory limit per sandbox container |
| `SANDBOX_CPU_LIMIT` | No | `0.5` | CPU limit per sandbox container |

---

## Health Check

```
GET /health → {"status":"ok"}
```

Use this endpoint for load balancer health checks and monitoring.

---

## Database Backups

### pg_dump

```bash
pg_dump -U aihub -h localhost aihub > backup_$(date +%Y%m%d).sql
```

### Automated daily backup (cron)

```bash
0 2 * * * pg_dump -U aihub aihub | gzip > /backups/aihub_$(date +\%Y\%m\%d).sql.gz
```

---

## Updating

```bash
cd Aihub
git pull
go build -o bin/aihub-server ./cmd/server
cd web && npm ci && npm run build && cd ..
sudo systemctl restart aihub
```

Migrations run automatically on server start — no separate migration step needed.
