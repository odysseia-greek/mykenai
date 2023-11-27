provider "google" {
  project = ""
  region  = "europe-west1"  # Change to your region if needed
}

resource "google_service_account" "vault_unseal_sa" {
  account_id   = ""  # Change to your desired account ID
  display_name = "Vault Unseal Service Account"
}

resource "google_service_account_key" "vault_unseal_sa_key" {
  service_account_id = google_service_account.vault_unseal_sa.id
}


resource "google_kms_key_ring_iam_binding" "key_ring_viewer" {
  key_ring_id = "${var.gcloud-project}/${var.keyring_location}/${var.key_ring}"
  role        = "roles/owner"

  members = [
    "serviceAccount:${google_service_account.vault_unseal_sa.email}",
  ]
}
