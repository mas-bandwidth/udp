# ----------------------------------------------------------------------------------------

variable "google_org_id" { type = string }
variable "google_billing_account" { type = string }
variable "google_location" { type = string }
variable "google_region" { type = string }
variable "google_zones" { type = list(string) }
variable "google_zone" { type = string }

variable "tag" { type = string }

# ----------------------------------------------------------------------------------------

terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0.0"
    }
  }
}

provider "google" {
  region      = var.google_region
  zone        = var.google_zone
}

# ----------------------------------------------------------------------------------------

resource "random_id" "postfix" {
  byte_length = 8
}

resource "google_project" "udp" {
  name            = "UDP Test"
  project_id      = "udp-${random_id.postfix.hex}"
  org_id          = var.google_org_id
  billing_account = var.google_billing_account
}

# ----------------------------------------------------------------------------------------

locals {
  services = [
    "compute.googleapis.com",                   # compute engine
    "storage.googleapis.com",                   # cloud storage
  ]
}

resource "google_project_service" "udp" {
  count    = length(local.services)
  project  = google_project.udp.project_id
  service  = local.services[count.index]
  timeouts {
    create = "30m"
    update = "40m"
  }
  disable_dependent_services = true
}

# ----------------------------------------------------------------------------------------

data "local_file" "client_go" {
  filename = "client.go"
}

data "local_file" "server_go" {
  filename = "server.go"
}

data "local_file" "backend_go" {
  filename = "backend.go"
}

data "local_file" "backend_service" {
  filename = "backend.service"
}

data "local_file" "go_mod" {
  filename = "go.mod"
}

data "archive_file" "source_zip" {
  type        = "zip"
  output_path = "source.zip"
  source {
    filename = "client.go"
    content  = data.local_file.client_go.content
  }
  source {
    filename = "server.go"
    content  = data.local_file.server_go.content
  }
  source {
    filename = "backend.go"
    content  = data.local_file.backend_go.content
  }
  source {
    filename = "backend.service"
    content  = data.local_file.backend_service.content
  }
  source {
    filename = "go.mod"
    content  = data.local_file.go_mod.content
  }
}

resource "google_storage_bucket" "source" {
  name          = "${var.google_org_id}_udp_source"
  project       = google_project.udp.project_id
  location      = "US"
  force_destroy = true
  public_access_prevention = "enforced"
  uniform_bucket_level_access = true
}

resource "google_storage_bucket_object" "source_zip" {
  name         = "source-${var.tag}.zip"
  source       = "source.zip"
  content_type = "application/zip"
  bucket       = google_storage_bucket.source.id
}

# ----------------------------------------------------------------------------------------

resource "google_service_account" "udp_runtime" {
  project  = google_project.udp.project_id
  account_id   = "udp-runtime"
  display_name = "UDP Runtime Service Account"
}

resource "google_project_iam_member" "udp_runtime_compute_viewer" {
  project = google_project.udp.project_id
  role    = "roles/compute.viewer"
  member  = google_service_account.udp_runtime.member
}

resource "google_storage_bucket_iam_member" "udp_runtime_storage_admin" {
  bucket = google_storage_bucket.source.name
  role   = "roles/storage.objectAdmin"
  member = google_service_account.udp_runtime.member
  depends_on = [google_storage_bucket.source]
}

# ----------------------------------------------------------------------------------------

resource "google_compute_network" "udp" {
  name                    = "udp"
  project                 = google_project.udp.project_id
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "udp" {
  name                     = "udp"
  project                  = google_project.udp.project_id
  ip_cidr_range            = "10.0.0.0/16"
  region                   = var.google_region
  network                  = google_compute_network.udp.id
  private_ip_google_access = true
}

resource "google_compute_firewall" "allow_ssh" {
  name          = "allow-ssh"
  project       = google_project.udp.project_id
  direction     = "INGRESS"
  network       = google_compute_network.udp.id
  source_ranges = ["130.211.0.0/22", "35.191.0.0/16", "35.235.240.0/20"]
  allow {
    protocol = "tcp"
    ports    = ["22"]
  }
  target_tags = ["allow-ssh"]
}

# ----------------------------------------------------------------------------------------

resource "google_compute_address" "client_address" {
  name    = "client-address"
  project = google_project.udp.project_id
}

resource "google_compute_instance" "client" {

  name         = "client-${var.tag}"
  project      = google_project.udp.project_id
  machine_type = "n1-standard-8"
  zone         = var.google_zone
  tags         = ["allow-ssh"]

  allow_stopping_for_update = true

  boot_disk {
    initialize_params {
      image = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    }
  }

  network_interface {
    network    = google_compute_network.udp.id
    subnetwork = google_compute_subnetwork.udp.id
    access_config {
      nat_ip = google_compute_address.client_address.address
    }
  }

  metadata = {
    startup-script = <<-EOF
    #!/bin/bash
    NEEDRESTART_SUSPEND=1 apt update -y
    NEEDRESTART_SUSPEND=1 apt upgrade -y
    NEEDRESTART_SUSPEND=1 apt install golang-go unzip -y
    mkdir /app
    cd /app
    gsutil cp gs://${var.google_org_id}_udp_source/source-${var.tag}.zip .
    unzip *.zip
    go get
    EOF
  }

  service_account {
    email  = google_service_account.udp_runtime.email
    scopes = ["cloud-platform"]
  }
}

# ----------------------------------------------------------------------------------------

resource "google_compute_address" "server_address" {
  name    = "server-address"
  project = google_project.udp.project_id
}

resource "google_compute_instance" "server" {

  name         = "server-${var.tag}"
  project      = google_project.udp.project_id
  machine_type = "n1-standard-8"
  zone         = var.google_zone
  tags         = ["allow-ssh"]

  allow_stopping_for_update = true

  boot_disk {
    initialize_params {
      image = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    }
  }

  network_interface {
    network    = google_compute_network.udp.id
    subnetwork = google_compute_subnetwork.udp.id
    access_config {
      nat_ip = google_compute_address.server_address.address
    }
  }

  metadata = {
    startup-script = <<-EOF
    #!/bin/bash
    NEEDRESTART_SUSPEND=1 apt update -y
    NEEDRESTART_SUSPEND=1 apt upgrade -y
    NEEDRESTART_SUSPEND=1 apt install golang-go unzip -y
    mkdir /app
    cd /app
    gsutil cp gs://${var.google_org_id}_udp_source/source-${var.tag}.zip .
    unzip *.zip
    go get
    EOF
  }

  service_account {
    email  = google_service_account.udp_runtime.email
    scopes = ["cloud-platform"]
  }
}

# ----------------------------------------------------------------------------------------

resource "google_compute_address" "backend_address" {
  name    = "backend-address"
  project = google_project.udp.project_id
}

resource "google_compute_instance" "backend" {

  name         = "backend-${var.tag}"
  project      = google_project.udp.project_id
  machine_type = "n1-standard-8"
  zone         = var.google_zone
  tags         = ["allow-ssh"]

  allow_stopping_for_update = true

  boot_disk {
    initialize_params {
      image = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    }
  }

  network_interface {
    network    = google_compute_network.udp.id
    subnetwork = google_compute_subnetwork.udp.id
    access_config {
      nat_ip = google_compute_address.backend_address.address
    }
  }

  metadata = {
    startup-script = <<-EOF
    #!/bin/bash
    NEEDRESTART_SUSPEND=1 apt update -y
    NEEDRESTART_SUSPEND=1 apt upgrade -y
    NEEDRESTART_SUSPEND=1 apt install golang-go unzip -y
    mkdir /app
    cd /app
    gsutil cp gs://${var.google_org_id}_udp_source/source-${var.tag}.zip .
    unzip *.zip
    go get
    cp backend.service /etc/systemd/system/backend.service
    systemctl daemon-reload
    systemctl start backend.service
    EOF
  }

  service_account {
    email  = google_service_account.udp_runtime.email
    scopes = ["cloud-platform"]
  }
}

# ----------------------------------------------------------------------------------------
