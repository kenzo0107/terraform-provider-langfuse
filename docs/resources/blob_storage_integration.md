---
page_title: "langfuse_blob_storage_integration Resource - terraform-provider-langfuse"
subcategory: ""
description: |-
  Manages a Langfuse blob storage integration for data export.
---

# langfuse_blob_storage_integration (Resource)

Manages a Langfuse blob storage integration. Blob storage integrations allow Langfuse to export trace data to external object storage (S3, GCS, or Azure Blob Storage).

> **Note:** `access_key_id` and `secret_access_key` are write-only and cannot be recovered via import.

## Example Usage

```terraform
resource "langfuse_blob_storage_integration" "s3" {
  type        = "S3"
  bucket_name = "my-langfuse-exports"
  region      = "us-east-1"
  prefix      = "langfuse/"

  access_key_id     = var.aws_access_key_id
  secret_access_key = var.aws_secret_access_key

  enabled = true
}
```

## Schema

### Required

- `type` (String) The storage backend type. Must be one of `S3`, `AZURE_BLOB`, or `GCS`.
- `bucket_name` (String) The name of the storage bucket.

### Optional

- `prefix` (String) An optional prefix for objects stored in the bucket.
- `region` (String) The region of the storage bucket (for S3-compatible storage).
- `endpoint` (String) A custom endpoint URL (for S3-compatible storage or self-hosted).
- `export_prefix` (String) An optional prefix used when exporting data.
- `access_key_id` (String, Sensitive) The access key ID. Write-only; not read back from the API.
- `secret_access_key` (String, Sensitive) The secret access key. Write-only; not read back from the API.
- `enabled` (Boolean) Whether the integration is enabled. Defaults to `true`.

### Read-Only

- `id` (String) The unique identifier of the blob storage integration.

## Import

Import is supported by integration ID. Note that credentials cannot be recovered after import:

```shell
terraform import langfuse_blob_storage_integration.s3 <integration_id>
```
