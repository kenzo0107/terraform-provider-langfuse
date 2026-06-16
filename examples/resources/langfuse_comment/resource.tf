resource "langfuse_comment" "example" {
  object_type = "TRACE"
  object_id   = "trace-id-here"
  content     = "This trace looks good."
}
