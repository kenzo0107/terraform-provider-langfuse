resource "langfuse_scim_user" "example" {
  user_name   = "jane.doe@example.com"
  email       = "jane.doe@example.com"
  given_name  = "Jane"
  family_name = "Doe"
  active      = true
  password    = var.initial_password
}
