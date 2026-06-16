---
page_title: "langfuse_custom_model Resource - terraform-provider-langfuse"
subcategory: ""
description: |-
  Manages a Langfuse custom model definition for cost tracking.
---

# langfuse_custom_model (Resource)

Manages a Langfuse custom model definition. Custom models allow you to define pricing for LLM models used in your application so Langfuse can calculate costs.

All attributes except `id` and `is_langfuse_managed` require replacement when changed.

## Example Usage

```terraform
resource "langfuse_custom_model" "example" {
  model_name    = "my-gpt4"
  match_pattern = "(?i)^my-gpt-4.*"
  unit          = "TOKENS"
  input_price   = 0.00003
  output_price  = 0.00006
}
```

## Schema

### Required

- `model_name` (String) The name of the custom model. Forces replacement when changed.
- `match_pattern` (String) The regex pattern to match model names in traces. Forces replacement when changed.

### Optional

- `unit` (String) The pricing unit (e.g. `TOKENS`). Forces replacement when changed.
- `input_price` (Number) Price per input token. Forces replacement when changed.
- `output_price` (Number) Price per output token. Forces replacement when changed.
- `total_price` (Number) Total price per token (alternative to input/output split). Forces replacement when changed.

### Read-Only

- `id` (String) The unique identifier of the custom model.
- `is_langfuse_managed` (Boolean) Whether this model is managed by Langfuse.

## Import

Import is supported using the following syntax:

```shell
terraform import langfuse_custom_model.example <model_id>
```
