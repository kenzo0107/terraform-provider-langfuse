---
page_title: "langfuse_llm_connection Resource - terraform-provider-langfuse"
subcategory: ""
description: |-
  Manages a Langfuse LLM connection for playground and evaluation features.
---

# langfuse_llm_connection (Resource)

Manages a Langfuse LLM connection. LLM connections configure external LLM providers used by the Langfuse playground and automated evaluation features.

> **Note:** `api_key` is write-only and will not be read back from the API. Changes to `name` require replacement.

## Example Usage

```terraform
resource "langfuse_llm_connection" "openai" {
  name                = "openai-production"
  provider            = "openai"
  api_key             = var.openai_api_key
  with_default_models = true
}
```

## Schema

### Required

- `name` (String) The name of the LLM connection. Forces replacement when changed.
- `provider` (String) The LLM provider (e.g. `openai`, `anthropic`, `azure`).

### Optional

- `base_url` (String) The base URL for the LLM provider API (for custom or Azure endpoints).
- `api_key` (String, Sensitive) The API key for the LLM provider. This is write-only and will not be read back from the API.
- `with_default_models` (Boolean) Whether to include default models for this provider. Defaults to the provider's setting.

### Read-Only

- `id` (String) The unique identifier of the LLM connection.

## Import

Import is supported by connection name (the API has no get-by-ID endpoint):

```shell
terraform import langfuse_llm_connection.openai openai-production
```
