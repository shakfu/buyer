# Docker Setup Guide

This guide explains how to deploy the Buyer application using Docker and Docker Compose with PostgreSQL and MinIO.

## Overview

The application can be deployed in two modes:

1. **Development Mode** (`docker-compose.dev.yml`): Single-node PostgreSQL and MinIO for local development
2. **Production Mode** (`docker-compose.prod.yml`): High-availability setup with distributed MinIO cluster

## Architecture

### Development Architecture
- **PostgreSQL**: Single instance for development
- **MinIO**: Single-node object storage
- **Buyer App**: Single instance with hot-reload support

### Production Architecture
- **PostgreSQL**: Single instance with replication support
- **MinIO**: 4-node distributed cluster with erasure coding
- **Buyer App**: Multiple replicas behind NGINX load balancer
- **NGINX**: Load balancer and reverse proxy
- **Prometheus**: Metrics and monitoring

## Prerequisites

- Docker Engine 20.10+
- Docker Compose 2.0+
- At least 4GB RAM for development
- At least 16GB RAM for production

## Development Setup

### 1. Configure Environment

```bash
# Copy the example environment file
cp .env.dev.example .env

# Edit .env with your preferred settings (optional)
nano .env
```

### 2. Start Services

```bash
# Start all services
docker-compose -f docker-compose.dev.yml up -d

# View logs
docker-compose -f docker-compose.dev.yml logs -f

# View specific service logs
docker-compose -f docker-compose.dev.yml logs -f buyer
```

### 3. Access Services

- **Buyer Application**: http://localhost:8080
- **MinIO Console**: http://localhost:9001 (minioadmin/minioadmin)
- **PostgreSQL**: localhost:5432 (buyer/buyerdev123)

### 4. Manage Services

```bash
# Stop services
docker-compose -f docker-compose.dev.yml down

# Stop and remove volumes (clean slate)
docker-compose -f docker-compose.dev.yml down -v

# Rebuild buyer image
docker-compose -f docker-compose.dev.yml build buyer
docker-compose -f docker-compose.dev.yml up -d buyer

# View service status
docker-compose -f docker-compose.dev.yml ps
```

## Production Setup

### 1. Prepare Configuration Files

#### PostgreSQL Configuration

```bash
mkdir -p postgres

# Create postgresql.conf
cat > postgres/postgresql.conf <<EOF
# PostgreSQL Production Configuration
max_connections = 200
shared_buffers = 2GB
effective_cache_size = 6GB
maintenance_work_mem = 512MB
checkpoint_completion_target = 0.9
wal_buffers = 16MB
default_statistics_target = 100
random_page_cost = 1.1
effective_io_concurrency = 200
work_mem = 5MB
min_wal_size = 1GB
max_wal_size = 4GB
max_worker_processes = 4
max_parallel_workers_per_gather = 2
max_parallel_workers = 4
max_parallel_maintenance_workers = 2
EOF

# Create pg_hba.conf
cat > postgres/pg_hba.conf <<EOF
# TYPE  DATABASE        USER            ADDRESS                 METHOD
local   all             all                                     trust
host    all             all             127.0.0.1/32            md5
host    all             all             ::1/128                 md5
host    all             all             0.0.0.0/0               md5
host    replication     replicator      0.0.0.0/0               md5
EOF
```

#### NGINX Configuration

```bash
mkdir -p nginx

cat > nginx/nginx.conf <<EOF
events {
    worker_connections 1024;
}

http {
    upstream buyer_backend {
        least_conn;
        server buyer:8080 max_fails=3 fail_timeout=30s;
    }

    server {
        listen 80;
        server_name _;

        location /health {
            access_log off;
            return 200 "healthy\n";
            add_header Content-Type text/plain;
        }

        location / {
            proxy_pass http://buyer_backend;
            proxy_set_header Host \$host;
            proxy_set_header X-Real-IP \$remote_addr;
            proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto \$scheme;

            # Timeouts
            proxy_connect_timeout 60s;
            proxy_send_timeout 60s;
            proxy_read_timeout 60s;
        }
    }

    # HTTPS configuration (uncomment and configure for production)
    # server {
    #     listen 443 ssl http2;
    #     server_name yourdomain.com;
    #
    #     ssl_certificate /etc/nginx/certs/public.crt;
    #     ssl_certificate_key /etc/nginx/certs/private.key;
    #     ssl_protocols TLSv1.2 TLSv1.3;
    #     ssl_ciphers HIGH:!aNULL:!MD5;
    #
    #     location / {
    #         proxy_pass http://buyer_backend;
    #         proxy_set_header Host \$host;
    #         proxy_set_header X-Real-IP \$remote_addr;
    #         proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
    #         proxy_set_header X-Forwarded-Proto \$scheme;
    #     }
    # }
}
EOF
```

#### Prometheus Configuration

```bash
mkdir -p prometheus

cat > prometheus/prometheus.yml <<EOF
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'minio'
    metrics_path: /minio/v2/metrics/cluster
    scheme: http
    static_configs:
      - targets: ['minio1:9000', 'minio2:9000', 'minio3:9000', 'minio4:9000']

  - job_name: 'buyer'
    static_configs:
      - targets: ['buyer:8080']
EOF
```

### 2. Generate TLS Certificates

```bash
mkdir -p certs

# Self-signed certificate (for testing)
openssl req -new -x509 -days 365 -nodes \
  -out certs/public.crt \
  -keyout certs/private.key \
  -subj "/C=US/ST=State/L=City/O=Organization/CN=buyer.yourdomain.com"

# For production, use Let's Encrypt or your organization's CA
```

### 3. Configure Environment

```bash
# Copy the production example
cp .env.prod.example .env

# Edit with STRONG passwords
nano .env

# IMPORTANT: Set strong passwords for:
# - POSTGRES_PASSWORD
# - POSTGRES_REPLICATION_PASSWORD
# - MINIO_ROOT_USER and MINIO_ROOT_PASSWORD (min 32 chars)
# - BUYER_USERNAME and BUYER_PASSWORD
```

### 4. Deploy Production Stack

```bash
# Start all services
docker-compose -f docker-compose.prod.yml up -d

# View logs
docker-compose -f docker-compose.prod.yml logs -f

# Check service health
docker-compose -f docker-compose.prod.yml ps

# Scale buyer application (optional)
docker-compose -f docker-compose.prod.yml up -d --scale buyer=5
```

### 5. Verify Deployment

```bash
# Check PostgreSQL
docker-compose -f docker-compose.prod.yml exec postgres \
  psql -U buyer -d buyer -c "SELECT version();"

# Check MinIO cluster status
docker-compose -f docker-compose.prod.yml exec mc \
  /usr/bin/mc admin info buyerprod

# Check buyer application
curl http://localhost/
```

## Database Management

### Backup Database

```bash
# Create backup
docker-compose -f docker-compose.prod.yml exec postgres \
  pg_dump -U buyer buyer > backup_$(date +%Y%m%d_%H%M%S).sql

# Backup with compression
docker-compose -f docker-compose.prod.yml exec postgres \
  pg_dump -U buyer buyer | gzip > backup_$(date +%Y%m%d_%H%M%S).sql.gz
```

### Restore Database

```bash
# Restore from backup
cat backup.sql | docker-compose -f docker-compose.prod.yml exec -T postgres \
  psql -U buyer buyer

# Restore from compressed backup
gunzip -c backup.sql.gz | docker-compose -f docker-compose.prod.yml exec -T postgres \
  psql -U buyer buyer
```

### Database Migrations

```bash
# Access PostgreSQL shell
docker-compose -f docker-compose.prod.yml exec postgres \
  psql -U buyer buyer

# Run migrations (if using migration tool)
docker-compose -f docker-compose.prod.yml exec buyer \
  ./buyer migrate
```

## MinIO Management

### Access MinIO Console

- **Development**: http://localhost:9001
- **Production**: http://localhost:9001 (configure NGINX for production access)

### Create Additional Buckets

```bash
# Using mc client
docker-compose -f docker-compose.prod.yml exec mc \
  /usr/bin/mc mb buyerprod/new-bucket

# Enable versioning
docker-compose -f docker-compose.prod.yml exec mc \
  /usr/bin/mc version enable buyerprod/new-bucket
```

### Backup MinIO Data

```bash
# Sync to another MinIO instance or S3
docker-compose -f docker-compose.prod.yml exec mc \
  /usr/bin/mc mirror buyerprod s3/backup-bucket
```

## Monitoring

### View Logs

```bash
# All services
docker-compose -f docker-compose.prod.yml logs -f

# Specific service
docker-compose -f docker-compose.prod.yml logs -f buyer

# Last 100 lines
docker-compose -f docker-compose.prod.yml logs --tail=100 buyer
```

### Prometheus Metrics

Access Prometheus at http://localhost:9090

Useful queries:
- MinIO capacity: `minio_cluster_capacity_usable_total_bytes`
- MinIO requests: `minio_s3_requests_total`
- Database connections: `pg_stat_database_numbackends`

### Health Checks

```bash
# Check all services
docker-compose -f docker-compose.prod.yml ps

# Check buyer health
curl http://localhost/

# Check MinIO health
curl http://localhost:9000/minio/health/live
```

## Troubleshooting

### Service Won't Start

```bash
# Check logs
docker-compose -f docker-compose.prod.yml logs buyer

# Check service status
docker-compose -f docker-compose.prod.yml ps

# Restart specific service
docker-compose -f docker-compose.prod.yml restart buyer
```

### Database Connection Issues

```bash
# Check PostgreSQL logs
docker-compose -f docker-compose.prod.yml logs postgres

# Test connection
docker-compose -f docker-compose.prod.yml exec postgres \
  psql -U buyer buyer -c "SELECT 1;"

# Check environment variables
docker-compose -f docker-compose.prod.yml exec buyer env | grep DB
```

### MinIO Cluster Issues

```bash
# Check cluster status
docker-compose -f docker-compose.prod.yml exec mc \
  /usr/bin/mc admin info buyerprod

# Check individual node health
curl http://localhost:9000/minio/health/cluster
```

### Reset Everything

```bash
# Stop and remove all containers and volumes
docker-compose -f docker-compose.prod.yml down -v

# Remove all data (WARNING: This deletes everything!)
docker volume prune -f

# Restart from scratch
docker-compose -f docker-compose.prod.yml up -d
```

## Security Recommendations

1. **Change Default Credentials**: Update all default passwords in `.env`
2. **Enable TLS/SSL**: Configure HTTPS for all services
3. **Firewall Rules**: Restrict access to internal networks
4. **Regular Updates**: Keep Docker images updated
5. **Secrets Management**: Use Docker secrets or external secret managers
6. **Network Isolation**: Use Docker networks to isolate services
7. **Regular Backups**: Automate database and MinIO backups
8. **Monitoring**: Set up alerting for failures and security events
9. **Access Logs**: Review and monitor access logs regularly
10. **Vulnerability Scanning**: Scan images for vulnerabilities

## Migration from SQLite

If you're migrating from SQLite to PostgreSQL:

1. **Export data from SQLite**:
   ```bash
   ./buyer export --format=csv --output=data.csv
   ```

2. **Start PostgreSQL stack**:
   ```bash
   docker-compose -f docker-compose.dev.yml up -d postgres
   ```

3. **Import data**:
   ```bash
   ./buyer import --format=csv --input=data.csv
   ```

## Performance Tuning

### PostgreSQL

- Adjust `shared_buffers` based on available RAM (25% of total RAM)
- Tune `work_mem` for complex queries
- Enable connection pooling (e.g., PgBouncer)
- Monitor slow queries with `pg_stat_statements`

### MinIO

- Use distributed mode for production (4+ nodes)
- Enable erasure coding for data redundancy
- Configure lifecycle policies for old data
- Use SSD storage for better performance

### Buyer Application

- Scale horizontally by increasing replicas
- Configure resource limits appropriately
- Enable caching where applicable
- Monitor memory and CPU usage

## Additional Resources

- [Docker Documentation](https://docs.docker.com/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [MinIO Documentation](https://min.io/docs/)
- [NGINX Documentation](https://nginx.org/en/docs/)
- [Prometheus Documentation](https://prometheus.io/docs/)
