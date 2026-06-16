resource "langfuse_blob_storage_integration" "s3" {
  type        = "S3"
  bucket_name = "my-langfuse-exports"
  region      = "us-east-1"
  prefix      = "langfuse/"

  access_key_id     = var.aws_access_key_id
  secret_access_key = var.aws_secret_access_key

  enabled = true
}

resource "langfuse_blob_storage_integration" "gcs" {
  type        = "GCS"
  bucket_name = "my-langfuse-gcs-bucket"
  enabled     = true
}
