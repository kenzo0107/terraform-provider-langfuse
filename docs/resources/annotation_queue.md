---
page_title: "langfuse_annotation_queue Resource - terraform-provider-langfuse"
subcategory: ""
description: |-
  Manages a Langfuse annotation queue.
---

# langfuse_annotation_queue (Resource)

Manages a Langfuse annotation queue. Annotation queues are used to route LLM outputs to human reviewers for scoring.

## Example Usage

```terraform
resource "langfuse_annotation_queue" "example" {
  name        = "review-queue"
  description = "Queue for human review of low-quality outputs"

  score_config_ids = [
    langfuse_score_config.quality.id,
  ]
}
```

## Schema

### Required

- `name` (String) The name of the annotation queue.

### Optional

- `description` (String) A description of the annotation queue.
- `score_config_ids` (List of String) List of score config IDs associated with this annotation queue.

### Read-Only

- `id` (String) The unique identifier of the annotation queue.
- `project_id` (String) The ID of the project this annotation queue belongs to.

## Import

Import is supported using the following syntax:

```shell
terraform import langfuse_annotation_queue.example <annotation_queue_id>
```
