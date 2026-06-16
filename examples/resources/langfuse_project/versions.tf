terraform {
  required_providers {
    langfuse = {
      source = "registry.terraform.io/kenzo0107/langfuse"
    }
  }
}

provider "langfuse" {
  # public_key = var.langfuse_public_key  # or LANGFUSE_PUBLIC_KEY env var
  # secret_key = var.langfuse_secret_key  # or LANGFUSE_SECRET_KEY env var
  # host       = "https://cloud.langfuse.com"  # or LANGFUSE_HOST env var
}
