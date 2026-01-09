variable "grafana_url" {
  description = "Grafana server URL"
  type        = string
  default     = "http://localhost:3000"
}

variable "grafana_auth" {
  description = "Grafana authentication (e.g., 'admin:admin' or use API key)"
  type        = string
  default     = "admin:admin"
  sensitive   = true
}

variable "loki_url" {
  description = "Loki server URL"
  type        = string
  default     = "http://loki:3100"
}
