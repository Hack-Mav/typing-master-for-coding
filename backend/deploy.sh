#!/bin/bash

# Deployment script for Google Cloud App Engine

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if gcloud is installed
if ! command -v gcloud &> /dev/null; then
    print_error "gcloud CLI is not installed. Please install it first."
    exit 1
fi

# Parse command line arguments
ENVIRONMENT=${1:-development}
PROJECT_ID=${2:-""}

if [ "$ENVIRONMENT" != "development" ] && [ "$ENVIRONMENT" != "production" ]; then
    print_error "Environment must be 'development' or 'production'"
    exit 1
fi

# Set project ID based on environment if not provided
if [ -z "$PROJECT_ID" ]; then
    if [ "$ENVIRONMENT" = "production" ]; then
        PROJECT_ID="typing-master-prod"
    else
        PROJECT_ID="typing-master-dev"
    fi
fi

print_status "Deploying to $ENVIRONMENT environment"
print_status "Project ID: $PROJECT_ID"

# Set the active project
gcloud config set project $PROJECT_ID

# Choose the appropriate app.yaml file
if [ "$ENVIRONMENT" = "production" ]; then
    APP_YAML="app.yaml"
else
    APP_YAML="app.dev.yaml"
fi

print_status "Using configuration file: $APP_YAML"

# Run tests before deployment
print_status "Running tests..."
go test ./... || {
    print_error "Tests failed. Deployment aborted."
    exit 1
}

# Build the application
print_status "Building application..."
go build -o main . || {
    print_error "Build failed. Deployment aborted."
    exit 1
}

# Deploy to App Engine
print_status "Deploying to Google Cloud App Engine..."
gcloud app deploy $APP_YAML --quiet || {
    print_error "Deployment failed."
    exit 1
}

# Get the deployed URL
APP_URL=$(gcloud app browse --no-launch-browser 2>/dev/null | grep -o 'https://[^[:space:]]*')

print_status "Deployment completed successfully!"
print_status "Application URL: $APP_URL"

# Optional: Open the application in browser
read -p "Do you want to open the application in your browser? (y/n): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    gcloud app browse
fi