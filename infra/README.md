# Bulbul Infrastructure Management

This directory contains the complete infrastructure as code (IaC) setup for the Bulbul ticketing system, using Terraform for resource provisioning and Ansible for configuration management and deployment.

## 🏗️ Architecture Overview

The infrastructure follows a multi-tier architecture pattern deployed on OpenStack cloud:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Load Balancer  │───▶│   API Servers   │───▶│   Data Layer    │
│     (Nginx)     │    │ (2+ instances)  │    │ PostgreSQL +    │
│                 │    │                 │    │ Valkey + ES     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │
         ▼
┌─────────────────┐
│   Monitoring    │
│ Prometheus +    │
│    Grafana      │
└─────────────────┘
```

## 📁 Directory Structure

```
infra/
├── terraform/           # Infrastructure provisioning
│   ├── main.tf         # Provider and base networking
│   ├── api.tf          # API server instances
│   ├── db.tf           # Database servers
│   ├── load_balancer.tf # Load balancer setup
│   ├── monitoring.tf   # Monitoring infrastructure
│   ├── variables.tf    # Terraform variables
│   ├── outputs.tf      # Output values
│   └── templates/      # Template files
├── roles/              # Ansible roles for services
│   ├── api_server/     # Go API application deployment
│   ├── postgresql/     # Database configuration
│   ├── valkey/         # Redis-compatible cache
│   ├── elasticsearch/  # Search engine setup
│   ├── load_balancer/  # Nginx load balancer
│   ├── monitoring/     # Prometheus + Grafana
│   └── node_exporter/  # System metrics collection
├── group_vars/         # Ansible variable files
├── secrets/            # Encrypted credentials
├── playbook.yml        # Main Ansible playbook
├── inventories.ini     # Server inventory
└── ansible.cfg         # Ansible configuration
```

## 🚀 Quick Start

### Prerequisites

1. **OpenStack credentials** - Set up environment variables for PSCloud
2. **SSH key pair** - Generate and configure SSH access
3. **Ansible Vault password** - Create `.ansible_password` file
4. **Terraform** >= 1.0
5. **Ansible** >= 2.9

### 1. SSH Configuration

Configure `~/.ssh/config` for secure access:

```bash
Host 192.168.0.*
  User ubuntu
  ProxyJump ubuntu@91.147.93.57
  Port 22
  IdentityFile ~/.ssh/id_ed25519_hackload
```

### 2. Environment Setup

Initialize environment variables from encrypted secrets:

```bash
cd infra
source ./init.sh
```

This script:
- Decrypts the environment file using Ansible Vault
- Exports OpenStack authentication variables
- Re-encrypts the secrets file for security

### 3. Infrastructure Provisioning

Deploy cloud resources using Terraform:

```bash
cd infra/terraform
terraform init
terraform plan    # Review planned changes
terraform apply
```

### 4. Service Configuration

Configure and deploy services using Ansible:

```bash
cd infra
ansible-playbook playbook.yml
```

## 🛠️ Infrastructure Components

### Terraform Resources

#### Networking
- **VPC**: Private network (192.168.0.0/24) with router and gateway
- **Security Groups**: Configured for HTTP/HTTPS, SSH, and internal traffic
- **Floating IPs**: Public IPs for load balancer and monitoring

#### Compute Instances
- **Load Balancer**: Single Nginx instance with SSL termination
- **API Servers**: Horizontally scalable Go application instances (default: 2)
- **Database Server**: PostgreSQL with persistent storage
- **Cache Server**: Valkey (Redis-compatible) for session management
- **Search Engine**: Elasticsearch for event search functionality
- **Monitoring**: Prometheus and Grafana for observability

### Ansible Roles

#### `api_server`
- Deploys compiled Go binaries (`api-server`, `sync-seats`)
- Configures systemd services with proper logging
- Sets up log rotation and monitoring
- **Templates**: Service configuration, environment files

#### `postgresql`
- Installs and configures PostgreSQL 14
- Creates application database and user
- Loads initial schema and data
- Configures backup and security settings
- **Templates**: `postgresql.conf`, `pg_hba.conf`, schema files

#### `valkey`
- Sets up Valkey (Redis fork) for caching
- Configures persistence and memory settings
- Loads user authentication data
- **Templates**: Configuration files, data loading scripts

#### `elasticsearch`
- Installs Elasticsearch 8.x
- Configures indexes for event search
- Sets up JVM options and security
- Loads sample event data for testing
- **Templates**: `elasticsearch.yml`, JVM options, data loading scripts

#### `load_balancer`
- Configures Nginx with SSL termination
- Sets up upstream API server pools
- Handles HTTP to HTTPS redirects
- Configures SSL certificates (Let's Encrypt ready)
- **Templates**: Nginx configuration, SSL setup scripts

#### `monitoring`
- Deploys Prometheus for metrics collection
- Configures Grafana with pre-built dashboards
- Sets up alerting rules and notification channels
- **Templates**: Prometheus config, Grafana dashboards

#### `node_exporter`
- Installs Node Exporter on all servers
- Collects system-level metrics
- Configures firewall rules for Prometheus scraping

## 🔐 Security & Secrets Management

### Ansible Vault
All sensitive data is encrypted using Ansible Vault:
- Database passwords
- API keys
- SSL certificates
- Service credentials

### Secret Files
```
secrets/
├── environment          # OpenStack and app environment variables
└── id_ed25519_hackload  # SSH private key for server access
```

### Variable Management
- **`group_vars/all.yaml`**: Global configuration flags
- **`group_vars/workloads.yaml`**: Encrypted application secrets
- **Role defaults**: Service-specific default values

## 📊 Monitoring & Observability

### Prometheus Metrics
- **Application metrics**: API response times, error rates, request counts
- **System metrics**: CPU, memory, disk, network utilization
- **Database metrics**: Connection pools, query performance
- **Cache metrics**: Hit rates, memory usage, eviction counts

### Grafana Dashboards
- **System Overview**: Infrastructure health and resource usage
- **Application Performance**: API metrics and business KPIs
- **Database Monitoring**: PostgreSQL performance and queries
- **Load Balancer Stats**: Request distribution and response codes

### Log Aggregation
- **Centralized logging**: All services log to structured formats
- **Log rotation**: Automated cleanup with 7-day retention
- **Error tracking**: Application errors forwarded to monitoring

## 🔧 Operational Procedures

### Scaling API Servers

1. **Update Terraform variables**:
   ```hcl
   variable "api_server_count" {
     default = 4  # Increase from 2
   }
   ```

2. **Apply infrastructure changes**:
   ```bash
   cd terraform && terraform apply
   ```

3. **Configure new instances**:
   ```bash
   cd .. && ansible-playbook playbook.yml -t api_server
   ```

### Deployment Process

1. **Build application**:
   ```bash
   cd .. && go build -o infra/roles/api_server/files/api-server cmd/api/main.go
   ```

2. **Deploy to servers**:
   ```bash
   cd infra && ansible-playbook playbook.yml -t api_server
   ```

### Database Management

- **Backup**: Automated daily backups with point-in-time recovery
- **Migration**: Schema changes applied through Ansible tasks
- **Monitoring**: Connection pooling and query performance tracking

### SSL Certificate Management

- **Automatic renewal**: Let's Encrypt integration with cron jobs
- **Certificate deployment**: Automated distribution to load balancers
- **Security headers**: HSTS, CSP, and other security configurations

## 🚨 Troubleshooting

### Common Issues

1. **Terraform state conflicts**: Use `terraform state pull` to inspect
2. **Ansible connectivity**: Verify SSH configuration and jump host
3. **Service failures**: Check systemd logs with `journalctl -u service-name`
4. **Database connections**: Monitor connection pools and tune settings

### Health Checks

All services expose health endpoints:
- **API**: `GET /health` - Application status
- **Database**: `GET /health/db` - Connection pool status  
- **Elasticsearch**: `GET /health/elasticsearch` - Search engine status
- **Prometheus**: `GET /-/healthy` - Metrics collection status

### Log Locations

```
/var/log/bulbul/api-server.log      # Application logs
/var/log/nginx/access.log           # Load balancer access
/var/log/postgresql/postgresql.log  # Database logs
/var/log/prometheus/prometheus.log  # Monitoring logs
```

## 📈 Performance Tuning

### Database Optimization
- Connection pooling configured for high concurrency
- Indexes optimized for search and analytics queries
- Prepared statements for common operations

### Cache Strategy
- Session data cached in Valkey for fast user authentication
- Event listings cached with intelligent invalidation
- Database query result caching for expensive operations

### Load Balancing
- Round-robin distribution across API servers
- Health check endpoints for automatic failover
- Session affinity for stateful operations

## 🔄 Backup & Recovery

### Automated Backups
- **Database**: Daily full backups with WAL archiving
- **Configuration**: All infrastructure code versioned in Git
- **Secrets**: Encrypted backups of Ansible Vault files

### Disaster Recovery
- **RTO**: 30 minutes for complete infrastructure recreation
- **RPO**: Maximum 1 hour of data loss from automated backups
- **Failover**: Automated health checks and load balancer routing

This infrastructure setup provides a robust, scalable, and maintainable foundation for the Bulbul ticketing system with comprehensive monitoring, security, and operational capabilities.