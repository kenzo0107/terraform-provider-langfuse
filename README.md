[![Go Reference](https://pkg.go.dev/badge/github.com/kenzo0107/terraform-provider-langfuse.svg)](https://pkg.go.dev/github.com/kenzo0107/terraform-provider-langfuse) [![Tests](https://github.com/kenzo0107/terraform-provider-langfuse/actions/workflows/test.yml/badge.svg)](https://github.com/kenzo0107/terraform-provider-langfuse/actions/workflows/test.yml)

# Terraform Provider Langfuse (Terraform Plugin Framework)

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.11
- [Go](https://golang.org/doc/install) >= 1.24.0

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

## Using the provider

```hcl
terraform {
  required_providers {
    langfuse = {
      source  = "kenzo0107/langfuse"
      version = "~> 0.1"
    }
  }
}

provider "langfuse" {
  public_key = var.langfuse_public_key
  secret_key = var.langfuse_secret_key
  # host = "https://cloud.langfuse.com" # optional, defaults to cloud.langfuse.com
}

resource "langfuse_project" "example" {
  name = "my-project"
}
```

See the [Terraform Registry documentation](https://registry.terraform.io/providers/kenzo0107/langfuse/latest/docs) for full usage.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of acceptance tests, run `make testacc`.

```shell
make testacc
```
