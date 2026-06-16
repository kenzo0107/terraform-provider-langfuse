resource "langfuse_score" "example" {
  name      = "quality"
  data_type = "NUMERIC"
  value     = 0.9
  trace_id  = "trace-id-here"
  comment   = "Automated quality score"
}
