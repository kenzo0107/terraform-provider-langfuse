resource "langfuse_evaluation_rule" "example" {
  name         = "trace-quality-check"
  target       = "TRACE"
  evaluator_id = langfuse_evaluator.example.id
  state        = "ACTIVE"
  sampling     = 0.5
}
