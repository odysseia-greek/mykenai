variable "gcloud-project" {
  description = "Google project name"
  default = "odysseia"
}

variable "gcloud-region" {
  default = "europe-west1"
}

variable "gcloud-zone" {
  default = "europe-west1"
}

variable "key_ring" {
  description = "Cloud KMS key ring name to create"
  default     = "autounseal"
}

variable "crypto_key" {
  default     = "vaultkey"
  description = "Crypto key name to create under the key ring"
}

variable "keyring_location" {
  default = "global"
}