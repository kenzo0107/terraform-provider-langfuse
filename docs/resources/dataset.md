---
page_title: "langfuse_dataset Resource - terraform-provider-langfuse"
subcategory: ""
description: |-
  Manages a Langfuse dataset.
---

# langfuse_dataset (Resource)

Manages a Langfuse dataset.

> **Note:** The Langfuse API does not support deleting datasets. Destroying this resource only removes it from Terraform state.

## Example Usage

```terraform
resource "langfuse_dataset" "example" {
  name        = "my-dataset"
  description = "A dataset for evaluating my application."
}
```

## Schema

### Required

- `name` (String) The name of the dataset. Changing this creates a new dataset.

### Optional

- `description` (String) An optional description for the dataset.

### Read-Only

- `id` (String) The unique identifier of the dataset.
- `project_id` (String) The project ID the dataset belongs to.

## Import

Import is supported by dataset name:

```shell
terraform import langfuse_dataset.example my-dataset
```
