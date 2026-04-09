output "bucket_id" {
  description = "S3 deployment bucket name"
  value       = aws_s3_bucket.deployment.id
}

output "bucket_arn" {
  description = "S3 deployment bucket ARN"
  value       = aws_s3_bucket.deployment.arn
}