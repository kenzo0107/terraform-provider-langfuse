resource "langfuse_annotation_queue" "example" {
  name        = "review-queue"
  description = "Queue for human review of low-quality outputs"

  score_config_ids = [
    langfuse_score_config.quality.id,
  ]
}
