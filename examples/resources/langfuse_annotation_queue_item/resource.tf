resource "langfuse_annotation_queue_item" "example" {
  queue_id = langfuse_annotation_queue.example.id
  trace_id = "trace-id-here"
  status   = "QUEUED"
}
