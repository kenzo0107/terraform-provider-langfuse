resource "langfuse_llm_connection" "openai" {
  name                = "openai-production"
  provider            = "openai"
  api_key             = var.openai_api_key
  with_default_models = true
}

resource "langfuse_llm_connection" "azure" {
  name     = "azure-openai"
  provider = "azure"
  base_url = "https://my-resource.openai.azure.com"
  api_key  = var.azure_api_key
}
