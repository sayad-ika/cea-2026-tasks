# ──────────────────────────────────────────────
# 1. BUILD + ZIP  (runs local PowerShell commands)
# ──────────────────────────────────────────────
resource "null_resource" "build_and_zip" {
  for_each = local.lambdas

  triggers = {
    source_hash = sha1(join("", [
      for f in fileset("${path.root}/../${each.value.cmd_path}", "**/*.go") :
      filesha1("${path.root}/../${each.value.cmd_path}/${f}")
    ]))
  }

  provisioner "local-exec" {
    working_dir = "${path.root}/.."
    interpreter = ["PowerShell", "-ExecutionPolicy", "Bypass", "-Command"]
    command     = <<-EOT
      Write-Host "Building lambda: ${each.key}"

      # Each lambda gets its own isolated output directory
      $outDir  = "./dist/${each.key}"
      $zipPath = "$outDir/${each.key}.zip"

      New-Item -ItemType Directory -Force -Path $outDir | Out-Null

      # Remove old artifacts so nothing is stale
      if (Test-Path "$outDir/bootstrap") { Remove-Item "$outDir/bootstrap" -Force }
      if (Test-Path $zipPath)            { Remove-Item $zipPath -Force }

      $env:GOOS        = "linux"
      $env:GOARCH      = "amd64"
      $env:CGO_ENABLED = "0"

      go build -o "$outDir/bootstrap" ${each.value.cmd_path}
      if ($LASTEXITCODE -ne 0) {
        Write-Error "Build failed for ${each.key}"
        exit 1
      }

      # Zip only the bootstrap binary (Lambda expects it named 'bootstrap')
      Compress-Archive -Path "$outDir/bootstrap" -DestinationPath $zipPath -Force
      Write-Host "Zipped to $zipPath"

      # Clean up binary, keep only the zip
      Remove-Item "$outDir/bootstrap" -Force
    EOT
  }
}

# ──────────────────────────────────────────────
# 2. UPLOAD ZIP TO S3
# ──────────────────────────────────────────────
resource "aws_s3_object" "lambda_zip" {
  for_each = local.lambdas

  bucket      = local.s3_bucket
  key         = "${local.s3_prefix}/${each.key}.zip"
  source      = "${path.root}/../dist/${each.key}/${each.key}.zip"  # <-- updated path
  source_hash = null_resource.build_and_zip[each.key].triggers.source_hash

  depends_on = [null_resource.build_and_zip]
}

# ──────────────────────────────────────────────
# 3. TELL LAMBDA TO PULL THE NEW ZIP FROM S3
# ──────────────────────────────────────────────
resource "null_resource" "update_lambda" {
  for_each = local.lambdas

  triggers = {
    s3_etag = aws_s3_object.lambda_zip[each.key].etag
  }

  provisioner "local-exec" {
    working_dir = "${path.root}/.."
    interpreter = ["PowerShell", "-ExecutionPolicy", "Bypass", "-Command"]
    command     = <<-EOT
      Write-Host "Updating Lambda function: ${each.value.lambda_function}"

      $zipPath = "./dist/${each.key}/${each.key}.zip"

      if (-not (Test-Path $zipPath)) {
        Write-Error "Zip not found at $zipPath"
        exit 1
      }

      aws lambda update-function-code `
        --function-name ${each.value.lambda_function} `
        --zip-file      fileb://$zipPath `
        --region        ${var.aws_region} | Out-Null

      if ($LASTEXITCODE -ne 0) {
        Write-Error "Lambda update failed for ${each.value.lambda_function}"
        exit 1
      }

      Write-Host "Waiting for update to complete..."
      aws lambda wait function-updated `
        --function-name ${each.value.lambda_function} `
        --region        ${var.aws_region}

      Write-Host "Done: ${each.value.lambda_function}"
    EOT
  }

  depends_on = [null_resource.build_and_zip]
}