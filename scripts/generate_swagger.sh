#!/bin/bash

# This script generates Swagger documentation from Go annotations

# Install swag if not already installed
if ! command -v swag &> /dev/null; then
    echo "Installing swag..."
    go install github.com/swaggo/swag/cmd/swag@v1.8.12
fi

# Generate Swagger documentation
echo "Generating Swagger documentation..."
swag init --dir ./cmd/auth,./internal/api/rest/handler,./internal/api/rest/router --output ./docs --generalInfo swagger.go

echo "Swagger documentation generated successfully!"