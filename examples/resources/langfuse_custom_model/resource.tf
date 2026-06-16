resource "langfuse_custom_model" "example" {
  model_name    = "my-gpt4"
  match_pattern = "(?i)^my-gpt-4.*"
  unit          = "TOKENS"
  input_price   = 0.00003
  output_price  = 0.00006
}
