# Grafana Dashboard Management with Terraform

Complete guide for managing Grafana dashboards using Infrastructure as Code (Terraform).

## Table of Contents

1. [Overview](#overview)
2. [Quick Start](#quick-start)
3. [What is Terraform?](#what-is-terraform)
4. [Local Development](#local-development)
5. [Production Deployment](#production-deployment)
6. [Dashboard Management](#dashboard-management)
7. [Advanced Features](#advanced-features)

---

## Overview

This project uses Terraform to manage Grafana dashboards, providing version control, automated deployments, and infrastructure as code capabilities.

### Directory Structure

```
hack/demo/
├── terraform/           # Terraform configuration
│   ├── main.tf         # Dashboard definition (inline HCL)
│   ├── variables.tf    # Configuration variables
│   ├── outputs.tf      # Outputs
│   └── README.md       # This file
```

### Key Benefits

- **Infrastructure as Code**: Dashboard defined in Terraform HCL
- **Version Control**: All changes tracked in Git
- **Code Review**: Dashboard changes go through PR process
- **Automated Deployment**: Deploy with `terraform apply`
- **Dynamic References**: Datasource UIDs automatically resolved
- **Multi-Environment**: Easy to deploy to dev/staging/prod
- **Rollback**: Easy to revert using Git

---

## Quick Start

### Prerequisites

1. Terraform installed (>= 1.0) - [Download](https://developer.hashicorp.com/terraform/install)
2. Grafana running and accessible
3. Loki datasource available

### Local Development - Fast Track

```bash
# 1. Start services
podman-compose up -d
# Grafana available at http://localhost:3000 (admin/admin)

# 2. Deploy dashboard
cd hack/demo/terraform
terraform init
terraform apply

# 3. View dashboard
# Visit http://localhost:3000/d/compliance-evidence
```

### Production Deployment - Fast Track

```bash
cd hack/demo/terraform

# Configure credentials
export TF_VAR_grafana_url="https://grafana.example.com"
export TF_VAR_grafana_auth="YOUR_API_KEY"
export TF_VAR_loki_url="http://loki:3100"

# Deploy
terraform init
terraform plan
terraform apply
```

### Basic Commands

```bash
# Initialize Terraform (first time only)
terraform init

# Preview changes
terraform plan

# Deploy the dashboard
terraform apply

# View state
terraform show

# Destroy resources (careful!)
terraform destroy
```

---

## What is Terraform?

Terraform is an **Infrastructure as Code (IaC)** tool that lets you manage infrastructure through code rather than manual processes.

### Why Use Terraform for Grafana Dashboards?

**Traditional Approach (Manual):**
```
┌─────────────────────────────────────────┐
│ 1. Open Grafana UI                      │
│ 2. Click "Create Dashboard"             │
│ 3. Add panels one by one                │
│ 4. Configure queries manually           │
│ 5. Save dashboard                       │
│                                         │
│ Problems:                                │
│ ❌ Hard to replicate                     │
│ ❌ No version control                    │
│ ❌ Manual errors                         │
│ ❌ Team members make conflicting changes │
│ ❌ Can't automate                        │
└─────────────────────────────────────────┘
```

**Terraform Approach (Automated):**
```
┌─────────────────────────────────────────┐
│ 1. Write dashboard config in HCL code  │
│ 2. Commit to Git                        │
│ 3. Run: terraform apply                 │
│                                         │
│ Benefits:                                │
│ ✅ Repeatable across environments        │
│ ✅ Version controlled in Git             │
│ ✅ Automated and consistent              │
│ ✅ Code review process                   │
│ ✅ CI/CD integration                     │
│ ✅ Dynamic datasource references         │
└─────────────────────────────────────────┘
```

### How Terraform Works

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Write      │     │   Preview    │     │    Apply     │
│   Config     │────▶│   Changes    │────▶│   Changes    │
│              │     │              │     │              │
│  main.tf     │     │ terraform    │     │  terraform   │
│              │     │   plan       │     │   apply      │
└──────────────┘     └──────────────┘     └──────────────┘
                            │
                            │ Shows what will change:
                            ▼
                     + Create dashboard
                     ~ Update panel query
                     - Delete old panel
```

**The Workflow:**

1. **Write** configuration in `main.tf`
2. **Initialize**: `terraform init` (downloads providers)
3. **Plan**: `terraform plan` (preview changes)
4. **Apply**: `terraform apply` (make changes)

---

## Local Development

### 1. Start Services with Podman Compose

```bash
# From the project root
podman-compose up -d

# Grafana will be available at http://localhost:3000
# Default credentials: admin/admin
```

**Note**: Dashboards are managed by Terraform, not auto-provisioned by Podman Compose.

### 2. Deploy Dashboard with Terraform

```bash
cd hack/demo/terraform

# Initialize (first time only)
terraform init

# Preview changes
terraform plan

# Deploy
terraform apply
```

The dashboard will be created at http://localhost:3000/d/compliance-evidence.

### 3. Making Changes

**Option 1: Edit Terraform HCL (Recommended)**

```bash
# 1. Edit the dashboard in main.tf
vim hack/demo/terraform/main.tf

# 2. Preview changes
terraform plan

# 3. Apply changes
terraform apply

# 4. Commit to Git
git add hack/demo/terraform/main.tf
git commit -m "Update dashboard configuration"
```

**Option 2: Export from Grafana UI**

1. Make changes in Grafana UI
2. Export dashboard JSON from Grafana
3. Convert JSON to HCL format
4. Update `main.tf`
5. Run `terraform apply`
6. Commit the changes to Git

---

## Production Deployment

### Prerequisites

1. Terraform installed (v1.0+)
2. Grafana instance running and accessible
3. Grafana API credentials (admin:password or API key)
4. Loki instance configured as data source

### Configuration Methods

**Method 1: Environment Variables (Recommended)**

```bash
export TF_VAR_grafana_url="https://grafana.example.com"
export TF_VAR_grafana_auth="YOUR_API_KEY"
export TF_VAR_loki_url="http://loki:3100"

terraform apply
```

**Method 2: terraform.tfvars File**

```bash
cat > terraform.tfvars <<EOF
grafana_url  = "https://grafana.example.com"
grafana_auth = "Bearer YOUR_API_KEY"
loki_url     = "http://loki:3100"
EOF

terraform apply
```

**Method 3: Command Line**

```bash
terraform apply \
  -var="grafana_url=https://grafana.example.com" \
  -var="grafana_auth=Bearer YOUR_API_KEY" \
  -var="loki_url=http://loki:3100"
```

### Using API Keys (Recommended)

Instead of username:password, use Grafana API keys for production:

1. **Create API key in Grafana:**
   - Go to Configuration → API Keys
   - Click "New API Key"
   - Set role to "Admin"
   - Copy the key

2. **Use the API key:**

```bash
export TF_VAR_grafana_auth="YOUR_API_KEY_HERE"
terraform apply
```

---

## Dashboard Management

### Inline HCL Approach

Dashboards are defined inline using HCL in `main.tf`:

```hcl
resource "grafana_dashboard" "compliance_evidence" {
  overwrite = true

  config_json = jsonencode({
    title = "Compliance Evidence Dashboard"
    uid   = "compliance-evidence"

    panels = [
      {
        id    = 1
        title = "Total Evidence Records"
        type  = "stat"
        datasource = {
          type = "loki"
          uid  = grafana_data_source.loki.uid  # Dynamic reference!
        }
        targets = [{
          expr = "sum(count_over_time({service_name=~\".+\"} [$__range]))"
        }]
      }
    ]
  })
}
```

**Benefits of Inline HCL:**

- **Dynamic datasource references**: Uses `grafana_data_source.loki.uid` for automatic UID resolution
- **Full version control**: Dashboard config is in HCL, part of your Terraform code
- **Infrastructure as Code**: Leverage Terraform variables, functions, and conditionals
- **Multi-environment support**: Easy to parameterize for different environments

### Dashboard Panels

The Compliance Evidence Dashboard includes 8 panels:

1. **Total Evidence Records** - Stat panel showing total evidence count
2. **Policy Evaluation Over Time** - Time series graph with stacked bars (Passed/Failed/Unknown)
3. **Evidence by Policy Engine** - Donut chart breakdown by policy engine
4. **Evidence by Policy Rule** - Donut chart breakdown by policy rule
5. **Evidence Count by Control (Real-time)** - Time series showing evidence rate per control
6. **Total Evidence Count by Control** - Stat panel showing total count per control
7. **Control Health: Evidence by Result** - Pie chart of evaluation results (Passed/Failed/Not Run/Needs Review)
8. **Control Health: Assessment Requirements and Evidence Count per Control ID** - Table showing control IDs with assessment requirements and evidence counts

### Drift Detection

To detect manual changes made in Grafana UI:

```bash
terraform plan
```

If you see changes, either:
- **Accept them**: Update `main.tf` with the changes and apply
- **Revert them**: Run `terraform apply` to restore Terraform's version
- **Refresh state**: Run `terraform refresh` to update state without applying

---

## Advanced Features

### 1. Managing Multiple Dashboards

Add more dashboard resources inline:

```hcl
resource "grafana_dashboard" "app_metrics" {
  overwrite = true

  config_json = jsonencode({
    title = "Application Metrics"
    uid   = "app-metrics"
    panels = [
      # Define panels here
    ]
  })
}

resource "grafana_dashboard" "infrastructure" {
  overwrite = true

  config_json = jsonencode({
    title = "Infrastructure Monitoring"
    uid   = "infrastructure"
    panels = [
      # Define panels here
    ]
  })
}
```

### 2. Folder Organization

Organize dashboards in folders:

```hcl
resource "grafana_folder" "compliance" {
  title = "Compliance Dashboards"
}

resource "grafana_dashboard" "compliance_evidence" {
  folder    = grafana_folder.compliance.id
  overwrite = true

  config_json = jsonencode({
    title = "Compliance Evidence"
    uid   = "compliance-evidence"
    panels = [
      # Panel definitions
    ]
  })
}
```

### 3. Environment-Specific Dashboards

```hcl
# variables.tf
variable "environment" {
  type = string
}

# main.tf
resource "grafana_dashboard" "app" {
  overwrite = true

  config_json = jsonencode({
    title = "Application Dashboard (${upper(var.environment)})"
    uid   = "app-${var.environment}"
    tags  = ["app", var.environment]

    panels = [
      {
        id    = 1
        title = "Requests (${var.environment})"
        type  = "graph"
        targets = [{
          expr = "rate(http_requests{env=\"${var.environment}\"}[5m])"
        }]
      }
    ]
  })
}
```

Usage:
```bash
# Development
terraform apply -var="environment=dev"

# Production
terraform apply -var="environment=prod"
```

### 4. Dynamic Datasource References

One of the biggest benefits of inline HCL dashboards:

```hcl
# Define datasource
resource "grafana_data_source" "loki" {
  type = "loki"
  name = "Loki"
  url  = var.loki_url
}

# Use in dashboard - automatically gets the correct UID!
resource "grafana_dashboard" "app" {
  overwrite = true

  config_json = jsonencode({
    title = "App Dashboard"
    panels = [
      {
        datasource = {
          type = "loki"
          uid  = grafana_data_source.loki.uid  # Dynamic!
        }
        targets = [{
          expr = "{app=\"myapp\"}"
        }]
      }
    ]
  })
}
```

If you recreate the datasource, the dashboard automatically gets the new UID!

---

## Additional Resources

- [Terraform Documentation](https://www.terraform.io/docs)
- [Grafana Terraform Provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs)
- [Infrastructure as Code Concepts](https://www.terraform.io/intro)

---
