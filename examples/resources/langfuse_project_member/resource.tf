resource "langfuse_project_member" "example" {
  project_id = langfuse_project.example.id
  user_id    = "user-id-here"
  role       = "MEMBER"
}
