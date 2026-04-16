# ──────────────────────────────────────────────
# 1. BUILD + ZIP  (bash / Linux-compatible)
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
    interpreter = ["/bin/bash", "-c"]
    command     = <<-EOT
      set -euo pipefail

      echo "Building lambda: ${each.key}"

      outDir="./dist/${each.key}"
      zipPath="$outDir/${each.key}.zip"

      mkdir -p "$outDir"

      rm -f "$outDir/bootstrap" "$zipPath"

      export GOOS=linux
      export GOARCH=amd64
      export CGO_ENABLED=0

      go build -o "$outDir/bootstrap" ${each.value.cmd_path}

      zip -j "$zipPath" "$outDir/bootstrap"
      echo "Zipped to $zipPath"

      rm -f "$outDir/bootstrap"
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
  source      = "${path.root}/../dist/${each.key}/${each.key}.zip" # <-- updated path
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
    interpreter = ["/bin/bash", "-c"]
    command     = <<-EOT
      set -euo pipefail

      echo "Updating Lambda function: ${each.value.lambda_function}"

      zipPath="./dist/${each.key}/${each.key}.zip"

      if [ ! -f "$zipPath" ]; then
        echo "Zip not found at $zipPath" >&2
        exit 1
      fi

      aws lambda update-function-code \
        --function-name ${each.value.lambda_function} \
        --zip-file      "fileb://$zipPath" \
        --region        ${var.aws_region} > /dev/null

      echo "Waiting for update to complete..."
      aws lambda wait function-updated \
        --function-name ${each.value.lambda_function} \
        --region        ${var.aws_region}

      echo "Done: ${each.value.lambda_function}"
    EOT
  }

  depends_on = [null_resource.build_and_zip]
}
