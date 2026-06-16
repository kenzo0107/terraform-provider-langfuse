---
page_title: "langfuse_scim_user Resource - terraform-provider-langfuse"
subcategory: ""
description: |-
  Manages a Langfuse SCIM user.
---

# langfuse_scim_user (Resource)

Manages a Langfuse SCIM user. Requires an organization-scoped API key with SCIM permissions.

> **Note:** `password` is write-only and cannot be recovered via import.

## Example Usage

```terraform
resource "langfuse_scim_user" "example" {
  user_name   = "jane.doe@example.com"
  email       = "jane.doe@example.com"
  given_name  = "Jane"
  family_name = "Doe"
  active      = true
  password    = var.initial_password
}
```

## Schema

### Required

- `user_name` (String) The username. Changing this creates a new user.
- `email` (String) The primary email address. Changing this creates a new user.

### Optional

- `given_name` (String) The given (first) name. Changing this creates a new user.
- `family_name` (String) The family (last) name. Changing this creates a new user.
- `active` (Boolean) Whether the user is active. Defaults to `true`.
- `external_id` (String) An external identifier. Changing this creates a new user.
- `password` (String, Sensitive) The initial password. Write-only; not read back from the API.

### Read-Only

- `id` (String) The unique identifier of the SCIM user.

## Import

Import is supported by user ID. Note that `password` cannot be recovered after import:

```shell
terraform import langfuse_scim_user.example <user_id>
```
