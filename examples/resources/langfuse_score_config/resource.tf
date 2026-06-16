resource "langfuse_score_config" "numeric" {
  name        = "quality"
  data_type   = "NUMERIC"
  min_value   = 0
  max_value   = 1
  description = "Output quality score between 0 and 1"
}

resource "langfuse_score_config" "categorical" {
  name      = "sentiment"
  data_type = "CATEGORICAL"

  categories = [
    { value = 0, label = "negative" },
    { value = 1, label = "neutral" },
    { value = 2, label = "positive" },
  ]
}
