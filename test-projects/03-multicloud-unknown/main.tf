# Mixed Azure + GCP estate with one resource type not yet in the portability
# dataset (aws_braket_quantum_task -> scored as the neutral unknown 0.50 and
# listed under "unscored" in the report). Exercises non-AWS scoring and the
# unknown-type path. Expected Terrasentry result: FAIL (repo ~0.51).

provider "aws" {
  region = "eu-west-2"
}

provider "azurerm" {
  features {}
}

provider "google" {
  project = "acme-prod"
}

resource "azurerm_linux_virtual_machine" "app" {
  name                = "acme-app"
  resource_group_name = "acme-rg"
  location            = "uksouth"
  size                = "Standard_B2s"
}

resource "azurerm_cosmosdb_account" "db" {
  name                = "acme-cosmos"
  resource_group_name = "acme-rg"
  location            = "uksouth"
  offer_type          = "Standard"
}

resource "google_storage_bucket" "assets" {
  name     = "acme-assets"
  location = "EU"
}

resource "google_spanner_instance" "ledger" {
  name         = "acme-ledger"
  config       = "regional-europe-west2"
  display_name = "acme-ledger"
  num_nodes    = 1
}

resource "google_container_cluster" "gke" {
  name     = "acme-gke"
  location = "europe-west2"
}

resource "aws_braket_quantum_task" "experiment" {
  device_arn = "arn:aws:braket:::device/quantum-simulator/amazon/sv1"
}
