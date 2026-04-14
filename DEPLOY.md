# Warungku — Deployment Guide

## Architecture Overview

```
Internet → Nginx (80/443) → Go Fiber App (:3000) → PostgreSQL (:5432, internal only)
```

Migrations run **automatically on every startup** — no manual step needed.

---

## Strategy 1: VPS / Cloud VM (Recommended for Indonesia)

Best for: full control, cheapest option. Good providers:
- **Hetzner Cloud** (CAX11, ~€4/mo — ARM, great value)
- **DigitalOcean** (Droplet $6/mo)
- **Vultr / Linode** are also fine
- **IDCloudHost / Niagahoster** if you want local Indonesian hosting

### Server requirements
- 1 vCPU, 1–2 GB RAM (Go is very lean)
- Ubuntu 22.04 LTS
- Ports 80 and 443 open

### First-time server setup

```bash
# Install Docker
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER

# Create app directory
sudo mkdir -p /opt/warungku
sudo chown $USER /opt/warungku
cd /opt/warungku

# Copy files: Dockerfile, docker-compose.yml, nginx.conf, .env.production
# (SCP from your machine or clone your repo)

# Rename and fill in secrets
cp .env.production .env
nano .env   # set POSTGRES_PASSWORD and JWT_SECRET

# Start everything
docker compose up -d

# Check logs
docker compose logs -f app
```

### SSL with Let's Encrypt (free HTTPS)

```bash
sudo apt install certbot
sudo certbot certonly --standalone -d yourdomain.com

# Certs land at:
# /etc/letsencrypt/live/yourdomain.com/fullchain.pem
# /etc/letsencrypt/live/yourdomain.com/privkey.pem

# Copy or symlink them into /opt/warungku/ssl/
mkdir ssl
ln -s /etc/letsencrypt/live/yourdomain.com/fullchain.pem ssl/fullchain.pem
ln -s /etc/letsencrypt/live/yourdomain.com/privkey.pem   ssl/privkey.pem

# Auto-renew (add to crontab)
0 3 * * * certbot renew --quiet && docker compose -f /opt/warungku/docker-compose.yml restart nginx
```

### Deploying updates

```bash
cd /opt/warungku
git pull                          # if using git on the server
docker compose build --no-cache app
docker compose up -d app
```

---

## Strategy 2: Railway (Easiest PaaS, free tier available)

1. Push your code to GitHub
2. Go to [railway.app](https://railway.app) → New Project → Deploy from GitHub
3. Add a **PostgreSQL** plugin — Railway auto-injects `DATABASE_URL`
4. Set env vars in the Railway dashboard:
   - `JWT_SECRET` = your secret
   - `APP_ENV` = production
5. Done — Railway builds your Dockerfile and deploys automatically on every push

**Cost**: ~$5/mo for hobby projects, free tier has sleep-mode.

---

## Strategy 3: Fly.io (Good free tier, global edge)

```bash
# Install flyctl
curl -L https://fly.io/install.sh | sh

# From your project root
fly launch    # detects Go app, auto-creates fly.toml
fly postgres create --name warungku-db
fly postgres attach warungku-db

# Set JWT secret
fly secrets set JWT_SECRET=your-secret-here APP_ENV=production

# Deploy
fly deploy
```

Fly auto-manages SSL, custom domains, and health checks.

---

## Strategy 4: Google Cloud Run (Serverless containers, pay-per-request)

Best for: apps with uneven/spiky traffic. Scales to zero.

```bash
# Build and push to Google Container Registry
gcloud builds submit --tag gcr.io/YOUR_PROJECT/warungku

# Deploy
gcloud run deploy warungku \
  --image gcr.io/YOUR_PROJECT/warungku \
  --platform managed \
  --region asia-southeast2 \
  --allow-unauthenticated \
  --set-env-vars "JWT_SECRET=...,APP_ENV=production" \
  --set-env-vars "DATABASE_URL=postgresql://..."
```

Use **Cloud SQL (PostgreSQL)** as the database and connect via the Cloud SQL Auth Proxy.

---

## CI/CD with GitHub Actions

Add these secrets to your GitHub repo (Settings → Secrets):

| Secret | Value |
|--------|-------|
| `VPS_HOST` | Your server IP |
| `VPS_USER` | `ubuntu` or `deploy` |
| `VPS_SSH_KEY` | Private SSH key (no passphrase) |

The workflow in `.github/workflows/deploy.yml` will:
1. Build a Docker image and push to GitHub Container Registry
2. SSH into your VPS and run `docker compose pull && docker compose up -d`

---

## Backup strategy

```bash
# Daily Postgres backup (add to server crontab)
0 2 * * * docker exec warungku_db pg_dump -U warungku warungku_prod \
  | gzip > /opt/backups/warungku_$(date +\%Y\%m\%d).sql.gz

# Keep 30 days
find /opt/backups -name "*.sql.gz" -mtime +30 -delete
```

---

## Quick recommendation for warungku

| Situation | Recommendation |
|-----------|----------------|
| Just getting started / testing | **Railway** — zero server management |
| Production, Indonesia users | **Hetzner or DigitalOcean VPS** + Docker |
| Need to scale later | **Fly.io** or **Cloud Run** |

For a warung management app with modest traffic, a **$6/mo DigitalOcean Droplet** with Docker Compose will comfortably handle hundreds of concurrent users and costs almost nothing.
