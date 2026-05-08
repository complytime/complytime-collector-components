variable "grafana_url" {
  description = "Grafana server URL"
  type        = string
  default     = "http://localhost:3000"
}

variable "grafana_auth" {
  description = "Grafana authentication (e.g., 'admin:secretpassword' or use API key)"
  type        = string
  sensitive   = true
  
  validation {
    condition     = var.grafana_auth != "admin:admin" && length(var.grafana_auth) > 0
    error_message = "Grafana authentication must be explicitly set and cannot use default 'admin:admin' credentials."
  }
}
