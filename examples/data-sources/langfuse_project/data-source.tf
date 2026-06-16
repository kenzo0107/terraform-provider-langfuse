data "langfuse_project" "example" {
  id = "clxxxxxxxxxxxxxxxxxxxxxxxx"
}

output "project_name" {
  value = data.langfuse_project.example.name
}
