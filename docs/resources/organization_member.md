---
page_title: "langfuse_organization_member Resource - terraform-provider-langfuse"
subcategory: ""
description: |-
  Manages a member's role within the Langfuse organization.
---

# langfuse_organization_member (Resource)

Manages a member's role within the Langfuse organization.

## Example Usage

```terraform
resource "langfuse_organization_member" "example" {
  user_id = "user-id-here"
  role    = "MEMBER"
}
```

## Schema

### Required

- `user_id` (String) The ID of the user to add to the organization. Forces replacement when changed.
- `role` (String) The role to assign to the user. Must be one of `OWNER`, `ADMIN`, `MEMBER`, or `VIEWER`.

### Read-Only

- `id` (String) The unique identifier of the membership (set to `user_id`).
- `email` (String) The email address of the member.
- `name` (String) The display name of the member.

## Import

Import is supported by user ID:

```shell
terraform import langfuse_organization_member.example <user_id>
```
