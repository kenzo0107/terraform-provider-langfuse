---
page_title: "langfuse_score Resource - terraform-provider-langfuse"
subcategory: ""
description: |-
  Manages a Langfuse score.
---

# langfuse_score (Resource)

Manages a Langfuse score. Scores attach numeric, boolean, categorical, or text evaluations to traces and observations. All attributes are immutable; any change requires replacement.

## Example Usage

```terraform
resource "langfuse_score" "example" {
  name      = "quality"
  data_type = "NUMERIC"
  value     = 0.9
  trace_id  = "trace-id-here"
  comment   = "Automated quality score"
}
```

## Schema

### Required

- `name` (String) The name of the score. Changing this creates a new score.

### Optional

- `value` (Number) The numeric value (for `NUMERIC` or `BOOLEAN` data types). Changing this creates a new score.
- `string_value` (String) The string value (for `CATEGORICAL`, `TEXT`, or `CORRECTION` data types). Changing this creates a new score.
- `data_type` (String) The data type. Must be `NUMERIC`, `BOOLEAN`, `CATEGORICAL`, `TEXT`, or `CORRECTION`. Changing this creates a new score.
- `trace_id` (String) The ID of the trace to associate with. Changing this creates a new score.
- `observation_id` (String) The ID of the observation to associate with. Changing this creates a new score.
- `config_id` (String) The ID of the score config. Changing this creates a new score.
- `comment` (String) An optional comment. Changing this creates a new score.
- `environment` (String) The environment. Changing this creates a new score.

### Read-Only

- `id` (String) The unique identifier of the score.
- `source` (String) The source of the score (set by Langfuse).

## Import

Import is supported by score ID:

```shell
terraform import langfuse_score.example <score_id>
```
