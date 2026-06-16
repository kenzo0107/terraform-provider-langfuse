---
page_title: "langfuse_score_config Resource - terraform-provider-langfuse"
subcategory: ""
description: |-
  Manages a Langfuse score configuration.
---

# langfuse_score_config (Resource)

Manages a Langfuse score configuration. Score configs define the evaluation criteria for LLM outputs.

> **Note:** The Langfuse API does not support deleting score configs. Destroying this resource will archive it instead.

## Example Usage

```terraform
resource "langfuse_score_config" "numeric" {
  name        = "quality"
  data_type   = "NUMERIC"
  min_value   = 0
  max_value   = 1
  description = "Output quality score between 0 and 1"
}

resource "langfuse_score_config" "categorical" {
  name      = "sentiment"
  data_type = "CATEGORICAL"

  categories = [
    { value = 0, label = "negative" },
    { value = 1, label = "neutral" },
    { value = 2, label = "positive" },
  ]
}
```

## Schema

### Required

- `name` (String) The name of the score config.
- `data_type` (String) The data type for scores. Must be one of `NUMERIC`, `BOOLEAN`, or `CATEGORICAL`. Forces replacement when changed.

### Optional

- `min_value` (Number) The minimum allowed value (for `NUMERIC` data type).
- `max_value` (Number) The maximum allowed value (for `NUMERIC` data type).
- `description` (String) A description of the score config.
- `categories` (List of Object) Category definitions (for `CATEGORICAL` data type). Each category has:
  - `value` (Number) The numeric value for this category.
  - `label` (String) The label for this category.

### Read-Only

- `id` (String) The unique identifier of the score config.
- `project_id` (String) The ID of the project this score config belongs to.
- `is_archived` (Boolean) Whether the score config is archived.

## Import

Import is supported using the following syntax:

```shell
terraform import langfuse_score_config.example <score_config_id>
```
