---
page_title: "langfuse_evaluation_rule Resource - terraform-provider-langfuse"
subcategory: ""
description: |-
  Manages a Langfuse evaluation rule.
---

# langfuse_evaluation_rule (Resource)

Manages a Langfuse evaluation rule.

> **Warning:** This resource uses the unstable Langfuse API (`/api/public/unstable/`) and may change without notice.

## Example Usage

```terraform
resource "langfuse_evaluation_rule" "example" {
  name         = "trace-quality-check"
  target       = "TRACE"
  evaluator_id = langfuse_evaluator.example.id
  state        = "ACTIVE"
  sampling     = 0.5
}
```

## Schema

### Required

- `name` (String) The name of the evaluation rule.
- `target` (String) The target type for the rule. Must be `TRACE` or `DATASET_RUN`. Changing this creates a new rule.
- `evaluator_id` (String) The ID of the evaluator to use. Changing this creates a new rule.

### Optional

- `state` (String) The state of the rule. Must be `ACTIVE` or `INACTIVE`.
- `filter` (String) A JSON string representing filter conditions for the rule.
- `mapping` (String) A JSON string representing variable mappings for the evaluator.
- `sampling` (Number) The sampling rate for the rule (0.0 to 1.0).
- `priority` (Number) The execution priority of the rule.

### Read-Only

- `id` (String) The unique identifier of the evaluation rule.

## Import

Import is supported by evaluation rule ID:

```shell
terraform import langfuse_evaluation_rule.example <rule_id>
```
