resource "langfuse_dataset_item" "example" {
  dataset_name    = langfuse_dataset.example.name
  input           = jsonencode({ question = "What is the capital of France?" })
  expected_output = jsonencode({ answer = "Paris" })
  status          = "ACTIVE"
}
