output "dashboard_url" {
  description = "URL to the compliance evidence dashboard"
  value       = "${var.grafana_url}/d/${grafana_dashboard.compliance_evidence.dashboard_id}"
}

output "loki_datasource_uid" {
  description = "UID of the Loki datasource"
  value       = grafana_data_source.loki.uid
}
