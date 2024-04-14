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

data "local_file" "client_service" {
  filename = "client.service"
}

data "local_file" "server_c" {
  filename = "server.c"
}

data "local_file" "server_xdp_c" {
  filename = "server_xdp.c"
}

data "local_file" "server_service" {
  filename = "server.service"
}

data "local_file" "go_mod" {
  filename = "go.mod"
}

data "local_file" "makefile" {
  filename = "Makefile"
}

data "archive_file" "source_zip" {
  type        = "zip"
  output_path = "source-${var.tag}.zip"
  source {
    filename = "client.go"
    content  = data.local_file.client_go.content
  }
  source {
    filename = "client.service"
    content  = data.local_file.client_service.content
  }
  source {
    filename = "server.c"
    content  = data.local_file.server_c.content
  }
  source {
    filename = "server_xdp.c"
    content  = data.local_file.server_xdp_c.content
  }
  source {
    filename = "server.service"
    content  = data.local_file.server_service.content
  }
  source {
    filename = "go.mod"
    content  = data.local_file.go_mod.content
  }
  source {
    filename = "Makefile"
    content  = data.local_file.makefile.content
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
  source       = "source-${var.tag}.zip"
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

resource "google_compute_router" "router" {
  name    = "router-to-internet"
  network = google_compute_network.udp.id
  project = google_project.udp.project_id
  region  = var.google_region
}

resource "google_compute_router_nat" "nat" {
  name                               = "nat"
  project                            = google_project.udp.project_id
  router                             = google_compute_router.router.name
  region                             = var.google_region
  nat_ip_allocate_option             = "AUTO_ONLY"
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"
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

resource "google_compute_firewall" "allow_http" {
  name          = "allow-http"
  project       = google_project.udp.project_id
  direction     = "INGRESS"
  network       = google_compute_network.udp.id
  source_ranges = ["0.0.0.0/0"]
  allow {
    protocol = "tcp"
    ports    = ["50000"]
  }
  target_tags = ["allow-http"]
}

resource "google_compute_firewall" "allow_udp" {
  name          = "allow-udp"
  project       = google_project.udp.project_id
  direction     = "INGRESS"
  network       = google_compute_network.udp.id
  source_ranges = ["0.0.0.0/0"]
  allow {
    protocol = "udp"
  }
  target_tags = ["allow-udp"]
}

# ----------------------------------------------------------------------------------------

resource "google_compute_instance_template" "client" {

  name         = "client-${var.tag}"

  project      = google_project.udp.project_id

  machine_type = "n1-standard-8"

  network_interface {
    network    = google_compute_network.udp.id
    subnetwork = google_compute_subnetwork.udp.id
    access_config {}
  }

  tags = ["allow-ssh"]

  disk {
    source_image = "ubuntu-os-cloud/ubuntu-minimal-2204-lts"
    auto_delete  = true
    boot         = true
    disk_type    = "pd-ssd"
  }

  metadata = {
    startup-script = <<-EOF2
    #!/bin/bash
    NEEDRESTART_SUSPEND=1 apt update -y
    NEEDRESTART_SUSPEND=1 apt upgrade -y
    NEEDRESTART_SUSPEND=1 apt install golang-go unzip -y
    mkdir /app
    cd /app
    gsutil cp gs://${var.google_org_id}_udp_source/source-${var.tag}.zip .
    unzip *.zip
    export HOME=/app
    go get
    go build client.go
    cat <<EOF > /app/client.env
    NUM_CLIENTS=1000
    SERVER_ADDRESS=${google_compute_instance.server.network_interface[0].network_ip}:40000
    EOF
    cat <<EOF > /etc/sysctl.conf
    net.core.rmem_max=1000000000
    net.core.wmem_max=1000000000
    net.core.netdev_max_backlog=10000
    EOF
    sysctl -p
    cp client.service /etc/systemd/system/client.service
    systemctl daemon-reload
    systemctl start client.service
    EOF2
  }

  service_account {
    email  = google_service_account.udp_runtime.email
    scopes = ["cloud-platform"]
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "google_compute_region_instance_group_manager" "client" {
  target_size               = 100
  name                      = "client"
  project                   = google_project.udp.project_id
  region                    = var.google_region
  distribution_policy_zones = var.google_zones
  version {
    instance_template       = google_compute_instance_template.client.id
    name                    = "primary"
  }
  base_instance_name        = "client"
  update_policy {
    type                           = "PROACTIVE"
    minimal_action                 = "REPLACE"
    most_disruptive_allowed_action = "REPLACE"
    max_surge_fixed                = 10
    max_unavailable_fixed          = 0
    replacement_method             = "SUBSTITUTE"
  }
  depends_on = [google_compute_instance_template.client]
}

# ----------------------------------------------------------------------------------------

resource "google_compute_address" "server_address" {
  name    = "server-${var.tag}-address"
  project = google_project.udp.project_id
}

resource "google_compute_instance" "server" {

  name         = "server-${var.tag}"
  project      = google_project.udp.project_id
  machine_type = "c3-highcpu-44"
  zone         = var.google_zone
  tags         = ["allow-ssh", "allow-udp"]

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

  advanced_machine_features {
    threads_per_core = 1
  }

  metadata = {

    startup-script = <<-EOF2

      #!/bin/bash

      NEEDRESTART_SUSPEND=1 apt autoremove -y
      NEEDRESTART_SUSPEND=1 apt update -y
      NEEDRESTART_SUSPEND=1 apt upgrade -y
      NEEDRESTART_SUSPEND=1 apt dist-upgrade -y
      NEEDRESTART_SUSPEND=1 apt full-upgrade -y
      NEEDRESTART_SUSPEND=1 apt install libcurl3-gnutls-dev build-essential vim golang-go wget libsodium-dev flex bison clang unzip libc6-dev-i386 gcc-12 dwarves libelf-dev pkg-config m4 libpcap-dev net-tools -y
      NEEDRESTART_SUSPEND=1 apt install linux-headers-`uname -r` linux-tools-`uname -r` -y
      NEEDRESTART_SUSPEND=1 apt autoremove -y

      sudo cp /sys/kernel/btf/vmlinux /usr/lib/modules/`uname -r`/build/

      mkdir /app
      cd /app
      gsutil cp gs://${var.google_org_id}_udp_source/source-${var.tag}.zip .
      unzip *.zip

      wget https://github.com/xdp-project/xdp-tools/releases/download/v1.4.2/xdp-tools-1.4.2.tar.gz
      tar -zxf xdp-tools-1.4.2.tar.gz
      cd xdp-tools-1.4.2
      ./configure
      make -j && make install

      cd lib/libbpf/src
      make -j && make install
      ldconfig

      cd /app
      make

      cat <<EOF > /etc/sysctl.conf
      net.core.rmem_max=1000000000
      net.core.wmem_max=1000000000
      net.core.netdev_max_backlog=10000
      EOF
      sysctl -p

      cp server.service /etc/systemd/system/server.service

      systemctl daemon-reload

      systemctl start server.service

    EOF2
  }

  service_account {
    email  = google_service_account.udp_runtime.email
    scopes = ["cloud-platform"]
  }
}

# ----------------------------------------------------------------------------------------
