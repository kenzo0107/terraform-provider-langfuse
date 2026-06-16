---
page_title: "langfuse_prompt Resource - terraform-provider-langfuse"
subcategory: ""
description: |-
  Manages a Langfuse prompt.
---

# langfuse_prompt (Resource)

Manages a Langfuse prompt. Prompts are versioned — each update to `text`, `messages`, `labels`, or `tags` creates a new version in Langfuse. Changes to `name` or `type` require replacement (delete all versions and recreate).

Destroying this resource deletes **all versions** of the prompt.

## Example Usage

```terraform
resource "langfuse_prompt" "summarize" {
  name   = "summarize"
  type   = "text"
  text   = "Summarize the following text in 3 sentences:\n\n{{text}}"
  labels = ["production"]
  tags   = ["summarization"]
}

resource "langfuse_prompt" "assistant" {
  name = "assistant"
  type = "chat"

  messages = [
    { role = "system", content = "You are a helpful assistant." },
    { role = "user", content = "{{user_input}}" },
  ]

  labels = ["production"]
}
```

## Schema

### Required

- `name` (String) The name of the prompt. Forces replacement when changed.
- `type` (String) The type of the prompt: `text` or `chat`. Forces replacement when changed.

### Optional

- `text` (String) The prompt text content (for `type = "text"`).
- `messages` (List of Object) Chat messages (for `type = "chat"`). Each message has:
  - `role` (String) The role of the message author (e.g. `user`, `assistant`, `system`).
  - `content` (String) The content of the message.
- `labels` (List of String) Labels to attach to the prompt version.
- `tags` (List of String) Tags to attach to the prompt.

### Read-Only

- `id` (String) The unique identifier of the prompt version (format: `{name}:v{version}`).
- `version` (Number) The version number assigned by Langfuse after creation.

## Import

Import is supported by prompt name (imports the latest version):

```shell
terraform import langfuse_prompt.summarize summarize
```
