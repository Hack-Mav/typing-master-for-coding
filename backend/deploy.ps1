# PowerShell deployment script for Google Cloud App Engine

param(
    [Parameter(Position=0)]
    [ValidateSet("development", "production")]
    [string]$Environment = "development",
    
    [Parameter(Position=1)]
    [string]$ProjectId = ""
)

# Function to write colored output
function Write-Status {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

# Check if gcloud is installed
if (-not (Get-Command gcloud -ErrorAction SilentlyContinue)) {
    Write-Error "gcloud CLI is not installed. Please install it first."
    exit 1
}

# Set project ID based on environment if not provided
if ([string]::IsNullOrEmpty($ProjectId)) {
    if ($Environment -eq "production") {
        $ProjectId = "typing-master-prod"
    } else {
        $ProjectId = "typing-master-dev"
    }
}

Write-Status "Deploying to $Environment environment"
Write-Status "Project ID: $ProjectId"

# Set the active project
gcloud config set project $ProjectId

# Choose the appropriate app.yaml file
if ($Environment -eq "production") {
    $AppYaml = "app.yaml"
} else {
    $AppYaml = "app.dev.yaml"
}

Write-Status "Using configuration file: $AppYaml"

# Run tests before deployment
Write-Status "Running tests..."
$testResult = go test ./...
if ($LASTEXITCODE -ne 0) {
    Write-Error "Tests failed. Deployment aborted."
    exit 1
}

# Build the application
Write-Status "Building application..."
$buildResult = go build -o main .
if ($LASTEXITCODE -ne 0) {
    Write-Error "Build failed. Deployment aborted."
    exit 1
}

# Deploy to App Engine
Write-Status "Deploying to Google Cloud App Engine..."
$deployResult = gcloud app deploy $AppYaml --quiet
if ($LASTEXITCODE -ne 0) {
    Write-Error "Deployment failed."
    exit 1
}

Write-Status "Deployment completed successfully!"

# Optional: Open the application in browser
$response = Read-Host "Do you want to open the application in your browser? (y/n)"
if ($response -eq "y" -or $response -eq "Y") {
    gcloud app browse
}