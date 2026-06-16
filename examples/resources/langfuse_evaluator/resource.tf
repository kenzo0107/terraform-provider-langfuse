resource "langfuse_evaluator" "example" {
  name = "llm-quality-evaluator"
  type = "llm_as_judge"

  prompt = jsonencode({
    model       = "gpt-4"
    system      = "You are an expert evaluator."
    temperature = 0
  })
}
