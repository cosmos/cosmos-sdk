#!/bin/bash

# =================================
# Tajeor Blockchain Explorer
# Professional Deployment Script
# =================================

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
DOMAIN=${DOMAIN:-"explorer.tajeor.network"}
EMAIL=${EMAIL:-"admin@tajeor.network"}
ENVIRONMENT=${ENVIRONMENT:-"production"}
COMPOSE_FILE="docker-compose.yml"
ENV_FILE=".env"

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_dependencies() {
    log_info "Checking dependencies..."
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    # Check Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
    
    # Check if Docker daemon is running
    if ! docker info &> /dev/null; then
        log_error "Docker daemon is not running. Please start Docker first."
        exit 1
    fi
    
    log_success "All dependencies are installed and running."
}

setup_environment() {
    log_info "Setting up environment..."
    
    # Create .env file from template if it doesn't exist
    if [[ ! -f "$ENV_FILE" ]]; then
        if [[ -f "environment.example" ]]; then
            cp environment.example "$ENV_FILE"
            log_info "Created .env file from template. Please edit it with your configuration."
        else
            log_error "No environment template found. Please create a .env file."
            exit 1
        fi
    fi
    
    # Generate secure passwords if not set
    if ! grep -q "DB_PASSWORD=" "$ENV_FILE" || grep -q "your_secure_database_password_here" "$ENV_FILE"; then
        DB_PASSWORD=$(openssl rand -base64 32)
        sed -i "s/your_secure_database_password_here/$DB_PASSWORD/" "$ENV_FILE"
        log_info "Generated secure database password."
    fi
    
    if ! grep -q "REDIS_PASSWORD=" "$ENV_FILE" || grep -q "your_secure_redis_password_here" "$ENV_FILE"; then
        REDIS_PASSWORD=$(openssl rand -base64 32)
        sed -i "s/your_secure_redis_password_here/$REDIS_PASSWORD/" "$ENV_FILE"
        log_info "Generated secure Redis password."
    fi
    
    if ! grep -q "JWT_SECRET=" "$ENV_FILE" || grep -q "your_jwt_secret_key_here" "$ENV_FILE"; then
        JWT_SECRET=$(openssl rand -base64 64)
        sed -i "s/your_jwt_secret_key_here_make_it_long_and_random/$JWT_SECRET/" "$ENV_FILE"
        log_info "Generated JWT secret."
    fi
    
    log_success "Environment setup completed."
}

setup_ssl() {
    log_info "Setting up SSL certificates..."
    
    # Create SSL directory
    mkdir -p ssl
    
    # Check if certificates already exist
    if [[ -f "ssl/fullchain.pem" && -f "ssl/privkey.pem" ]]; then
        log_info "SSL certificates already exist."
        return
    fi
    
    # For development, create self-signed certificates
    if [[ "$ENVIRONMENT" == "development" ]]; then
        log_warning "Creating self-signed certificates for development..."
        openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
            -keyout ssl/privkey.pem \
            -out ssl/fullchain.pem \
            -subj "/C=US/ST=State/L=City/O=Organization/CN=$DOMAIN"
        
        # Create DH parameters
        openssl dhparam -out ssl/dhparam.pem 2048
        
        log_success "Self-signed certificates created."
    else
        log_warning "For production, please obtain SSL certificates from a trusted CA."
        log_info "You can use Let's Encrypt with certbot:"
        log_info "certbot certonly --standalone -d $DOMAIN -m $EMAIL --agree-tos"
        log_info "Then copy the certificates to the ssl/ directory."
        
        # Create placeholder certificates for now
        touch ssl/fullchain.pem ssl/privkey.pem ssl/dhparam.pem
    fi
}

create_networks() {
    log_info "Creating Docker networks..."
    
    # Create custom network if it doesn't exist
    if ! docker network ls | grep -q "tajeor-network"; then
        docker network create tajeor-network
        log_success "Created tajeor-network."
    fi
}

setup_directories() {
    log_info "Setting up directory structure..."
    
    # Create necessary directories
    mkdir -p logs backups static
    mkdir -p monitoring/grafana/provisioning/{dashboards,datasources}
    mkdir -p nginx/sites-enabled
    
    # Set proper permissions
    chmod 755 logs backups static
    chmod 755 monitoring/grafana
    
    log_success "Directory structure created."
}

build_images() {
    log_info "Building Docker images..."
    
    # Build the main application image
    docker-compose build tajeor-explorer
    
    log_success "Docker images built successfully."
}

start_services() {
    log_info "Starting services..."
    
    # Start infrastructure services first
    docker-compose up -d postgres redis
    
    # Wait for database to be ready
    log_info "Waiting for database to be ready..."
    timeout=60
    while ! docker-compose exec -T postgres pg_isready -U tajeor &> /dev/null; do
        sleep 2
        timeout=$((timeout - 2))
        if [[ $timeout -le 0 ]]; then
            log_error "Database failed to start within 60 seconds."
            exit 1
        fi
    done
    
    # Start application services
    docker-compose up -d tajeor-explorer
    
    # Start monitoring services
    docker-compose up -d prometheus grafana
    
    # Start logging services
    docker-compose up -d elasticsearch kibana logstash
    
    # Start reverse proxy last
    docker-compose up -d nginx
    
    log_success "All services started successfully."
}

run_health_checks() {
    log_info "Running health checks..."
    
    # Check if services are healthy
    services=("postgres" "redis" "tajeor-explorer" "nginx")
    
    for service in "${services[@]}"; do
        if docker-compose ps "$service" | grep -q "Up (healthy)"; then
            log_success "$service is healthy."
        else
            log_warning "$service may not be healthy. Check with: docker-compose logs $service"
        fi
    done
    
    # Test API endpoint
    if curl -f http://localhost/api/health &> /dev/null; then
        log_success "API health check passed."
    else
        log_warning "API health check failed. Service may still be starting."
    fi
}

show_info() {
    log_info "Deployment completed successfully!"
    echo
    echo "🌐 Access your Tajeor Blockchain Explorer at:"
    echo "   Main site: https://$DOMAIN"
    echo "   API: https://$DOMAIN/api/"
    echo "   Health: https://$DOMAIN/health"
    echo
    echo "📊 Monitoring dashboards:"
    echo "   Grafana: http://localhost:3001"
    echo "   Prometheus: http://localhost:9090"
    echo "   Kibana: http://localhost:5601"
    echo
    echo "🔧 Management commands:"
    echo "   View logs: docker-compose logs -f [service]"
    echo "   Stop all: docker-compose down"
    echo "   Restart: docker-compose restart [service]"
    echo "   Update: ./scripts/deploy.sh --update"
    echo
    echo "📁 Important directories:"
    echo "   Logs: ./logs/"
    echo "   Backups: ./backups/"
    echo "   SSL: ./ssl/"
    echo
}

update_deployment() {
    log_info "Updating deployment..."
    
    # Pull latest images
    docker-compose pull
    
    # Rebuild application image
    docker-compose build tajeor-explorer
    
    # Restart services with zero downtime
    docker-compose up -d --force-recreate tajeor-explorer
    
    log_success "Deployment updated successfully."
}

cleanup() {
    log_info "Cleaning up old Docker resources..."
    
    # Remove unused images
    docker image prune -f
    
    # Remove unused volumes
    docker volume prune -f
    
    # Remove unused networks
    docker network prune -f
    
    log_success "Cleanup completed."
}

backup_data() {
    log_info "Creating backup..."
    
    # Create timestamp
    timestamp=$(date +"%Y%m%d_%H%M%S")
    backup_dir="backups/backup_$timestamp"
    
    mkdir -p "$backup_dir"
    
    # Backup database
    docker-compose exec -T postgres pg_dump -U tajeor tajeor_explorer > "$backup_dir/database.sql"
    
    # Backup configuration
    cp -r ssl "$backup_dir/"
    cp .env "$backup_dir/"
    cp docker-compose.yml "$backup_dir/"
    
    # Create archive
    tar -czf "backups/tajeor_explorer_backup_$timestamp.tar.gz" -C backups "backup_$timestamp"
    rm -rf "$backup_dir"
    
    log_success "Backup created: tajeor_explorer_backup_$timestamp.tar.gz"
}

# Main deployment function
main() {
    echo "🚀 Tajeor Blockchain Explorer - Professional Deployment"
    echo "======================================================"
    echo
    
    case "${1:-deploy}" in
        "deploy")
            check_dependencies
            setup_environment
            setup_ssl
            create_networks
            setup_directories
            build_images
            start_services
            sleep 30  # Give services time to start
            run_health_checks
            show_info
            ;;
        "update")
            update_deployment
            ;;
        "backup")
            backup_data
            ;;
        "cleanup")
            cleanup
            ;;
        "stop")
            log_info "Stopping all services..."
            docker-compose down
            log_success "All services stopped."
            ;;
        "restart")
            log_info "Restarting all services..."
            docker-compose restart
            log_success "All services restarted."
            ;;
        "logs")
            docker-compose logs -f "${2:-tajeor-explorer}"
            ;;
        "status")
            docker-compose ps
            ;;
        *)
            echo "Usage: $0 {deploy|update|backup|cleanup|stop|restart|logs|status}"
            echo
            echo "Commands:"
            echo "  deploy  - Full deployment (default)"
            echo "  update  - Update existing deployment"
            echo "  backup  - Create backup"
            echo "  cleanup - Clean up Docker resources"
            echo "  stop    - Stop all services"
            echo "  restart - Restart all services"
            echo "  logs    - View logs (optionally specify service)"
            echo "  status  - Show service status"
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@" 