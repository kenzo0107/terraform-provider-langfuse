---
page_title: "langfuse_annotation_queue_item Resource - terraform-provider-langfuse"
subcategory: ""
description: |-
  Manages a Langfuse annotation queue item.
---

# langfuse_annotation_queue_item (Resource)

Manages a Langfuse annotation queue item (a trace or observation added to an annotation queue for human review).

## Example Usage

```terraform
resource "langfuse_annotation_queue_item" "example" {
  queue_id = langfuse_annotation_queue.example.id
  trace_id = "trace-id-here"
  status   = "QUEUED"
}
```

## Schema

### Required

- `queue_id` (String) The ID of the annotation queue. Changing this creates a new item.
- `trace_id` (String) The ID of the trace to annotate. Changing this creates a new item.

### Optional

- `observation_id` (String) The ID of the observation to annotate. Changing this creates a new item.
- `status` (String) The status of the queue item. Must be `QUEUED`, `ACTIVE`, or `COMPLETED`.

### Read-Only

- `id` (String) The unique identifier of the queue item.

## Import

Import is supported by `{queue_id}/{item_id}`:

```shell
terraform import langfuse_annotation_queue_item.example <queue_id>/<item_id>
```
