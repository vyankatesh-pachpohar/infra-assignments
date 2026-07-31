terraform {
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.30"
    }
  }
}

provider "kubernetes" {
  config_path = "~/.kube/config"
}

variable "namespace" {
  default = "config-service"
}

variable "db_user" {
  default = "postgres"
}

variable "db_password" {
  default   = "postgres"
  sensitive = true
}

variable "db_name" {
  default = "configdb"
}

resource "kubernetes_namespace" "this" {
  metadata {
    name = var.namespace
  }
}

resource "kubernetes_config_map" "app_config" {
  metadata {
    name      = "config-service-configmap"
    namespace = kubernetes_namespace.this.metadata[0].name
  }
  data = {
    APP_PORT  = "8080"
    DB_HOST   = "config-service-postgres"
    DB_PORT   = "5432"
    DB_NAME   = var.db_name
  }
}

resource "kubernetes_secret" "db_credentials" {
  metadata {
    name      = "config-service-db-secret"
    namespace = kubernetes_namespace.this.metadata[0].name
  }
  data = {
    DB_USER     = var.db_user
    DB_PASSWORD = var.db_password
  }
  type = "Opaque"
}
