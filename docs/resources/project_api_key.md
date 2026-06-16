---
page_title: "langfuse_project_api_key Resource - terraform-provider-langfuse"
subcategory: ""
description: |-
  Manages a Langfuse project API key.
---

# langfuse_project_api_key (Resource)

Manages a Langfuse project API key. The `secret_key` is only available immediately after creation and is stored in Terraform state. If the state is lost, the secret key cannot be recovered and a new key must be created.

> **Warning:** Import is not supported because the secret key cannot be recovered after the initial creation.

## Example Usage

```terraform
resource "langfuse_project_api_key" "example" {
  project_id = langfuse_project.example.id
  note       = "CI/CD pipeline key"
}

output "langfuse_secret_key" {
  value     = langfuse_project_api_key.example.secret_key
  sensitive = true
}
```

## Schema

### Required

- `project_id` (String) The ID of the project this API key belongs to. Forces replacement when changed.

### Optional

- `note` (String) An optional note to identify the API key. Forces replacement when changed.

### Read-Only

- `id` (String) The unique identifier of the API key.
- `public_key` (String) The public key assigned by Langfuse.
- `secret_key` (String, Sensitive) The secret key. Only available immediately after creation.
- `display_secret_key` (String) A partially masked representation of the secret key.
