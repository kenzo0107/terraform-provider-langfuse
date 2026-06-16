resource "langfuse_prompt" "text_example" {
  name   = "summarize"
  type   = "text"
  text   = "Summarize the following text in 3 sentences:\n\n{{text}}"
  labels = ["production"]
  tags   = ["summarization"]
}

resource "langfuse_prompt" "chat_example" {
  name = "assistant"
  type = "chat"

  messages = [
    { role = "system", content = "You are a helpful assistant." },
    { role = "user", content = "{{user_input}}" },
  ]

  labels = ["production"]
}
