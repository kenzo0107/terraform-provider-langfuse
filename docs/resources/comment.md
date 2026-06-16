---
page_title: "langfuse_comment Resource - terraform-provider-langfuse"
subcategory: ""
description: |-
  Manages a Langfuse comment.
---

# langfuse_comment (Resource)

Manages a Langfuse comment on a trace, observation, session, or prompt.

> **Note:** The Langfuse API does not support updating or deleting comments. Destroying this resource only removes it from Terraform state.

## Example Usage

```terraform
resource "langfuse_comment" "example" {
  object_type = "TRACE"
  object_id   = "trace-id-here"
  content     = "This trace looks good."
}
```

## Schema

### Required

- `object_type` (String) The type of object to comment on. Must be `TRACE`, `OBSERVATION`, `SESSION`, or `PROMPT`. Changing this creates a new comment.
- `object_id` (String) The ID of the object. Changing this creates a new comment.
- `content` (String) The text content of the comment. Changing this creates a new comment.

### Optional

- `author_user_id` (String) The ID of the author. Changing this creates a new comment.

### Read-Only

- `id` (String) The unique identifier of the comment.
- `project_id` (String) The project ID the comment belongs to.

## Import

Import is supported by comment ID:

```shell
terraform import langfuse_comment.example <comment_id>
```
