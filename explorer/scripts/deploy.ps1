# =================================
# Tajeor Blockchain Explorer
# Windows PowerShell Deployment Script
# =================================

param(
    [Parameter(Position = 0)]
    [ValidateSet("deploy", "update", "backup", "cleanup", "stop", "restart", "logs", "status")]
    [string]$Action = "deploy",
    
    [Parameter(Position = 1)]
    [string]$Service = "tajeor-explorer",
    
    [string]$Environment = "production",
    [string]$Domain = "explorer.tajeor.network",
    [string]$Email = "admin@tajeor.network"
)

# Colors for output
$Colors = @{
    Red    = "Red"
    Green  = "Green"
    Yellow = "Yellow"
    Blue   = "Blue"
    Cyan   = "Cyan"
}

# Configuration
$ComposeFile = "docker-compose.yml"
$EnvFile = ".env"

# Functions
function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor $Colors.Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor $Colors.Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor $Colors.Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor $Colors.Red
}

function Test-Dependencies {
    Write-Info "Checking dependencies..."
    
    # Check Docker
    try {
        $dockerVersion = docker --version
        Write-Success "Docker found: $dockerVersion"
    }
    catch {
        Write-Error "Docker is not installed or not in PATH. Please install Docker Desktop for Windows."
        exit 1
    }
    
    # Check Docker Compose
    try {
        $composeVersion = docker-compose --version
        Write-Success "Docker Compose found: $composeVersion"
    }
    catch {
        Write-Error "Docker Compose is not available. Please ensure Docker Desktop is properly installed."
        exit 1
    }
    
    # Check if Docker daemon is running
    try {
        docker info | Out-Null
        Write-Success "Docker daemon is running."
    }
    catch {
        Write-Error "Docker daemon is not running. Please start Docker Desktop."
        exit 1
    }
    
    # Check Node.js (optional)
    try {
        $nodeVersion = node --version
        Write-Info "Node.js found: $nodeVersion"
    }
    catch {
        Write-Warning "Node.js not found. Some development features may not work."
    }
}

function Initialize-Environment {
    Write-Info "Setting up environment..."
    
    # Create .env file from template if it doesn't exist
    if (-not (Test-Path $EnvFile)) {
        if (Test-Path "environment.example") {
            Copy-Item "environment.example" $EnvFile
            Write-Info "Created .env file from template. Please edit it with your configuration."
        }
        else {
            Write-Error "No environment template found. Please create a .env file."
            exit 1
        }
    }
    
    # Generate secure passwords if needed
    $envContent = Get-Content $EnvFile -Raw
    
    if ($envContent -match "your_secure_database_password_here") {
        $dbPassword = [System.Web.Security.Membership]::GeneratePassword(32, 0)
        $envContent = $envContent -replace "your_secure_database_password_here", $dbPassword
        Write-Info "Generated secure database password."
    }
    
    if ($envContent -match "your_secure_redis_password_here") {
        $redisPassword = [System.Web.Security.Membership]::GeneratePassword(32, 0)
        $envContent = $envContent -replace "your_secure_redis_password_here", $redisPassword
        Write-Info "Generated secure Redis password."
    }
    
    if ($envContent -match "your_jwt_secret_key_here_make_it_long_and_random") {
        $jwtSecret = [System.Web.Security.Membership]::GeneratePassword(64, 0)
        $envContent = $envContent -replace "your_jwt_secret_key_here_make_it_long_and_random", $jwtSecret
        Write-Info "Generated JWT secret."
    }
    
    # Update domain settings
    $envContent = $envContent -replace "explorer\.tajeor\.network", $Domain
    
    Set-Content $EnvFile $envContent
    Write-Success "Environment setup completed."
}

function Initialize-SSL {
    Write-Info "Setting up SSL certificates..."
    
    # Create SSL directory
    if (-not (Test-Path "ssl")) {
        New-Item -ItemType Directory -Path "ssl" | Out-Null
    }
    
    # Check if certificates already exist
    if ((Test-Path "ssl/fullchain.pem") -and (Test-Path "ssl/privkey.pem")) {
        Write-Info "SSL certificates already exist."
        return
    }
    
    # For development on Windows, create self-signed certificates
    if ($Environment -eq "development") {
        Write-Warning "Creating self-signed certificates for development..."
        
        # Use OpenSSL if available, otherwise use PowerShell certificates
        try {
            & openssl req -x509 -nodes -days 365 -newkey rsa:2048 `
                -keyout ssl/privkey.pem `
                -out ssl/fullchain.pem `
                -subj "/C=US/ST=State/L=City/O=Organization/CN=$Domain"
            
            & openssl dhparam -out ssl/dhparam.pem 2048
            Write-Success "Self-signed certificates created with OpenSSL."
        }
        catch {
            Write-Warning "OpenSSL not found. Creating PowerShell certificates..."
            
            # Create self-signed certificate using PowerShell
            $cert = New-SelfSignedCertificate -DnsName $Domain -CertStoreLocation "cert:\LocalMachine\My" -KeyLength 2048
            
            # Export certificate
            $certPath = "ssl/fullchain.pem"
            $keyPath = "ssl/privkey.pem"
            
            # Export as PEM (simplified)
            $certBase64 = [Convert]::ToBase64String($cert.RawData)
            $pemCert = "-----BEGIN CERTIFICATE-----`n$certBase64`n-----END CERTIFICATE-----"
            Set-Content $certPath $pemCert
            
            # Create dummy private key (not recommended for production)
            Set-Content $keyPath "-----BEGIN PRIVATE KEY-----`nDUMMY_KEY_FOR_DEVELOPMENT`n-----END PRIVATE KEY-----"
            Set-Content "ssl/dhparam.pem" "-----BEGIN DH PARAMETERS-----`nDUMMY_DH_PARAMS`n-----END DH PARAMETERS-----"
            
            Write-Success "PowerShell certificates created."
        }
    }
    else {
        Write-Warning "For production, please obtain SSL certificates from a trusted CA."
        Write-Info "You can use Let's Encrypt or purchase certificates from a CA."
        
        # Create placeholder certificates
        New-Item -ItemType File -Path "ssl/fullchain.pem" -Force | Out-Null
        New-Item -ItemType File -Path "ssl/privkey.pem" -Force | Out-Null
        New-Item -ItemType File -Path "ssl/dhparam.pem" -Force | Out-Null
    }
}

function Initialize-Directories {
    Write-Info "Setting up directory structure..."
    
    # Create necessary directories
    $directories = @(
        "logs", "backups", "static",
        "monitoring/grafana/provisioning/dashboards",
        "monitoring/grafana/provisioning/datasources",
        "nginx/sites-enabled"
    )
    
    foreach ($dir in $directories) {
        if (-not (Test-Path $dir)) {
            New-Item -ItemType Directory -Path $dir -Force | Out-Null
        }
    }
    
    Write-Success "Directory structure created."
}

function Build-Images {
    Write-Info "Building Docker images..."
    
    try {
        docker-compose build tajeor-explorer
        Write-Success "Docker images built successfully."
    }
    catch {
        Write-Error "Failed to build Docker images: $_"
        exit 1
    }
}

function Start-Services {
    Write-Info "Starting services..."
    
    try {
        # Start infrastructure services first
        Write-Info "Starting database and cache services..."
        docker-compose up -d postgres redis
        
        # Wait for database to be ready
        Write-Info "Waiting for database to be ready..."
        $timeout = 60
        do {
            Start-Sleep -Seconds 2
            $timeout -= 2
            try {
                docker-compose exec -T postgres pg_isready -U tajeor | Out-Null
                $dbReady = $true
            }
            catch {
                $dbReady = $false
            }
        } while (-not $dbReady -and $timeout -gt 0)
        
        if ($timeout -le 0) {
            Write-Error "Database failed to start within 60 seconds."
            exit 1
        }
        
        # Start application services
        Write-Info "Starting application services..."
        docker-compose up -d tajeor-explorer
        
        # Start monitoring services
        Write-Info "Starting monitoring services..."
        docker-compose up -d prometheus grafana
        
        # Start logging services
        Write-Info "Starting logging services..."
        docker-compose up -d elasticsearch kibana logstash
        
        # Start reverse proxy last
        Write-Info "Starting reverse proxy..."
        docker-compose up -d nginx
        
        Write-Success "All services started successfully."
    }
    catch {
        Write-Error "Failed to start services: $_"
        exit 1
    }
}

function Test-HealthChecks {
    Write-Info "Running health checks..."
    
    # Check if services are healthy
    $services = @("postgres", "redis", "tajeor-explorer", "nginx")
    
    foreach ($service in $services) {
        try {
            $status = docker-compose ps $service
            if ($status -match "Up.*healthy") {
                Write-Success "$service is healthy."
            }
            else {
                Write-Warning "$service may not be healthy. Check with: docker-compose logs $service"
            }
        }
        catch {
            Write-Warning "Could not check status of $service"
        }
    }
    
    # Test API endpoint
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:3000/api/health" -UseBasicParsing -TimeoutSec 10
        if ($response.StatusCode -eq 200) {
            Write-Success "API health check passed."
        }
    }
    catch {
        Write-Warning "API health check failed. Service may still be starting."
    }
}

function Show-Info {
    Write-Info "Deployment completed successfully!"
    Write-Host ""
    Write-Host "🌐 Access your Tajeor Blockchain Explorer at:" -ForegroundColor $Colors.Cyan
    Write-Host "   Main site: http://localhost (or https://$Domain)" -ForegroundColor White
    Write-Host "   API: http://localhost:3000/api/" -ForegroundColor White
    Write-Host "   Health: http://localhost:3000/api/health" -ForegroundColor White
    Write-Host ""
    Write-Host "📊 Monitoring dashboards:" -ForegroundColor $Colors.Cyan
    Write-Host "   Grafana: http://localhost:3001" -ForegroundColor White
    Write-Host "   Prometheus: http://localhost:9090" -ForegroundColor White
    Write-Host "   Kibana: http://localhost:5601" -ForegroundColor White
    Write-Host ""
    Write-Host "🔧 Management commands:" -ForegroundColor $Colors.Cyan
    Write-Host "   View logs: docker-compose logs -f [service]" -ForegroundColor White
    Write-Host "   Stop all: docker-compose down" -ForegroundColor White
    Write-Host "   Restart: docker-compose restart [service]" -ForegroundColor White
    Write-Host "   Update: .\scripts\deploy.ps1 -Action update" -ForegroundColor White
    Write-Host ""
    Write-Host "📁 Important directories:" -ForegroundColor $Colors.Cyan
    Write-Host "   Logs: .\logs\" -ForegroundColor White
    Write-Host "   Backups: .\backups\" -ForegroundColor White
    Write-Host "   SSL: .\ssl\" -ForegroundColor White
    Write-Host ""
}

function Update-Deployment {
    Write-Info "Updating deployment..."
    
    try {
        # Pull latest images
        docker-compose pull
        
        # Rebuild application image
        docker-compose build tajeor-explorer
        
        # Restart services with zero downtime
        docker-compose up -d --force-recreate tajeor-explorer
        
        Write-Success "Deployment updated successfully."
    }
    catch {
        Write-Error "Failed to update deployment: $_"
        exit 1
    }
}

function Invoke-Cleanup {
    Write-Info "Cleaning up old Docker resources..."
    
    try {
        # Remove unused images
        docker image prune -f
        
        # Remove unused volumes
        docker volume prune -f
        
        # Remove unused networks
        docker network prune -f
        
        Write-Success "Cleanup completed."
    }
    catch {
        Write-Error "Cleanup failed: $_"
    }
}

function New-Backup {
    Write-Info "Creating backup..."
    
    try {
        # Create timestamp
        $timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
        $backupDir = "backups\backup_$timestamp"
        
        New-Item -ItemType Directory -Path $backupDir -Force | Out-Null
        
        # Backup database
        docker-compose exec -T postgres pg_dump -U tajeor tajeor_explorer | Out-File "$backupDir\database.sql"
        
        # Backup configuration
        Copy-Item -Recurse "ssl" "$backupDir\"
        Copy-Item ".env" "$backupDir\"
        Copy-Item "docker-compose.yml" "$backupDir\"
        
        # Create archive
        Compress-Archive -Path "$backupDir\*" -DestinationPath "backups\tajeor_explorer_backup_$timestamp.zip"
        Remove-Item $backupDir -Recurse -Force
        
        Write-Success "Backup created: tajeor_explorer_backup_$timestamp.zip"
    }
    catch {
        Write-Error "Backup failed: $_"
    }
}

# Add System.Web assembly for password generation
Add-Type -AssemblyName System.Web

# Main execution
function Main {
    Write-Host "🚀 Tajeor Blockchain Explorer - Professional Deployment (Windows)" -ForegroundColor $Colors.Cyan
    Write-Host "=================================================================" -ForegroundColor $Colors.Cyan
    Write-Host ""
    
    switch ($Action) {
        "deploy" {
            Test-Dependencies
            Initialize-Environment
            Initialize-SSL
            Initialize-Directories
            Build-Images
            Start-Services
            Start-Sleep -Seconds 30  # Give services time to start
            Test-HealthChecks
            Show-Info
        }
        "update" {
            Update-Deployment
        }
        "backup" {
            New-Backup
        }
        "cleanup" {
            Invoke-Cleanup
        }
        "stop" {
            Write-Info "Stopping all services..."
            docker-compose down
            Write-Success "All services stopped."
        }
        "restart" {
            Write-Info "Restarting all services..."
            docker-compose restart
            Write-Success "All services restarted."
        }
        "logs" {
            docker-compose logs -f $Service
        }
        "status" {
            docker-compose ps
        }
        default {
            Write-Host "Usage: .\deploy.ps1 -Action {deploy|update|backup|cleanup|stop|restart|logs|status}" -ForegroundColor $Colors.Yellow
            Write-Host ""
            Write-Host "Commands:" -ForegroundColor $Colors.Cyan
            Write-Host "  deploy  - Full deployment (default)" -ForegroundColor White
            Write-Host "  update  - Update existing deployment" -ForegroundColor White
            Write-Host "  backup  - Create backup" -ForegroundColor White
            Write-Host "  cleanup - Clean up Docker resources" -ForegroundColor White
            Write-Host "  stop    - Stop all services" -ForegroundColor White
            Write-Host "  restart - Restart all services" -ForegroundColor White
            Write-Host "  logs    - View logs (specify -Service)" -ForegroundColor White
            Write-Host "  status  - Show service status" -ForegroundColor White
            exit 1
        }
    }
}

# Run main function
Main 