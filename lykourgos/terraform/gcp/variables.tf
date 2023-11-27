variable "gcloud-project" {
  description = "Google project name"
}

variable "gcloud-region" {
  default = "europe-west1"
}

variable "gcloud-zone" {
  default = "europe-west1"
}

variable "key_ring" {
  description = "Cloud KMS key ring name to create"
  default     = ""
}

variable "crypto_key" {
  default     = ""
  description = "Crypto key name to create under the key ring"
}

variable "keyring_location" {
  default = "global"
}