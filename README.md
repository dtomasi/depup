# depup - Dependency Version Updater

> **_NOTE:_**  This project is in the early stages of development.
> Use with caution and always back up your files before running `depup`.
> Feedback and contributions are welcome!

[![Test](https://github.com/dtomasi/depup/actions/workflows/test.yml/badge.svg)](https://github.com/dtomasi/depup/actions/workflows/test.yml)
[![Release](https://github.com/dtomasi/depup/actions/workflows/release.yml/badge.svg)](https://github.com/dtomasi/depup/actions/workflows/release.yml)

## Overview

`depup` is a command-line tool designed to update dependency versions across configuration files.
It automatically detects and updates version references in YAML files, making it easy to keep dependencies consistent across your projects.

## Features

- **Annotated Updates**: Uses special comments to identify update targets (`# depup package=name`)
- **Multiple Configuration Formats**:
    - YAML files (`.yaml`, `.yml`) for Docker Compose, Kubernetes manifests, etc.
    - HCL files (`.tf`, `.tfvars`, `.hcl`) for Terraform configurations
    - Support for both inline and preceding line dependency comments
    - Works with different comment styles in HCL (`#` and `//`)
- **Recursive Directory Scanning**: Process entire directory structures with a single command
- **Dry Run Mode**: Preview changes before applying them
- **Configurable File Extensions**: Focus on specific file types
- **Quote Style Preservation**: Maintains the original quote style (single, double, or no quotes)
- **Cross-Platform Support**: Works on Linux, macOS, and Windows

## Installation

### From Binary Releases

Download the latest binary for your platform from the [Releases page](https://github.com/dtomasi/depup/releases).

### From Source

```bash
git clone https://github.com/dtomasi/depup.git
cd depup
go build -o depup
```

## How It Works

depup uses special comments in your configuration files to identify dependency declarations. When a file is processed:

1. It scans for comments in the format `# depup package=name`
2. It updates the version on the line following the comment
3. It preserves the original quote style (single, double, or no quotes)

## Usage

### YAML File Examples

#### Example 1: Kubernetes Deployment YAML

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: web-app
  template:
    spec:
      containers:
      # depup package=my-app
      - image: company/my-app:1.0.0
        name: web-app
        ports:
        - containerPort: 8080
```

Update the version with:

```bash
depup update deployment.yaml --package my-app=2.0.0
```

#### Example 2: Docker Compose YAML

```yaml
version: '3'
services:
  app:
    image: my-app:1.0.0 # depup package=my-app
    ports:
      - "8080:8080"
  redis:
    # depup package=redis
    image: redis:6.0.0
    ports:
      - "6379:6379"
```

Update multiple packages with:

```bash
depup update docker-compose.yaml --package my-app=2.0.0 --package redis=6.2.0
```

### HCL File Examples

#### Example 1: Terraform Provider Version

```hcl
terraform {
  required_providers {
    aws = {
      source = "hashicorp/aws"
      # depup package=aws-provider
      version = "4.0.0"
    }
  }
}

provider "aws" {
  region = "us-west-2"
}
```

Update the AWS provider version:

```bash
depup update main.tf --package aws-provider=4.5.0
```

#### Example 2: Terraform Module Version

```hcl
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  // depup package=vpc-module
  version = "3.14.0"

  name = "my-vpc"
  cidr = "10.0.0.0/16"
}
```

Update multiple Terraform files at once:

```bash
depup update terraform/ --package vpc-module=3.19.0 --ext .tf
```

By default, depup will recursively scan all directories. Use `--ext` flag to limit to specific file extensions.

## Development

### Requirements

- Go 1.21 or higher
- golangci-lint (for code quality checks)

## License

MIT
