# Build all CraftsBite Lambda functions for ARM64 (Graviton)
# Output: dist/{lambda_name}/{lambda_name}.zip
# Usage: .\scripts\build.ps1

$ErrorActionPreference = "Stop"

$lambdas = @("router", "gchat-router", "self", "management", "ops")
$env:GOOS = "linux"
$env:GOARCH = "arm64"
$env:CGO_ENABLED = "0"

Write-Host "=== Building CraftsBite Lambda functions (GOOS=linux, GOARCH=arm64) ===" -ForegroundColor Cyan

foreach ($lambda in $lambdas) {
    Write-Host ""
    Write-Host "--- Building $lambda ---" -ForegroundColor Yellow

    $srcDir = "./cmd/$lambda"

    if (-not (Test-Path $srcDir)) {
        Write-Host "ERROR: Source directory $srcDir not found" -ForegroundColor Red
        exit 1
    }

    # Create output directory
    New-Item -ItemType Directory -Force -Path "dist/$lambda" | Out-Null

    # Build the binary
    Write-Host "Compiling..."
    go build -o "dist/$lambda/bootstrap" "./$srcDir"

    if ($LASTEXITCODE -ne 0) {
        Write-Host "ERROR: Build failed for $lambda" -ForegroundColor Red
        exit 1
    }

    # Create zip (remove old zip first)
    $zipPath = "dist/$lambda/$lambda.zip"
    if (Test-Path $zipPath) {
        Remove-Item $zipPath
    }

    Push-Location "dist/$lambda"
    Compress-Archive -Path "bootstrap" -DestinationPath "$lambda.zip" -Force
    Pop-Location

    # Clean up binary
    Remove-Item "dist/$lambda/bootstrap" -ErrorAction SilentlyContinue

    Write-Host "OK: dist/$lambda/$lambda.zip" -ForegroundColor Green
}

Write-Host ""
Write-Host "=== All builds complete ===" -ForegroundColor Cyan
Write-Host "Run 'terraform -chdir=terraform apply' to deploy" -ForegroundColor Cyan