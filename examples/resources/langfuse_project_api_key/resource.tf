resource "langfuse_project_api_key" "example" {
  project_id = langfuse_project.example.id
  note       = "CI/CD pipeline key"
}

output "langfuse_secret_key" {
  value     = langfuse_project_api_key.example.secret_key
  sensitive = true
}
