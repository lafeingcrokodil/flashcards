# Infrastructure

## One-off setup

The project and the bucket for Terraform's state have to exist before
Terraform can run, so they're created by hand rather than by Terraform. These
steps are only needed for a new project.

1. Create the project and link it to the billing account (`gcloud billing
   accounts list` shows the account ID):

   ```sh
   gcloud projects create lafeingcrokodil-flashcards --name="<project_name>"
   ```

   ```sh
   gcloud billing projects link lafeingcrokodil-flashcards --billing-account=<billing_account_id>
   ```

1. Create the bucket for Terraform's state, with versioning so that earlier
   states can be recovered. Its name is the one in the `backend` block in
   `main.tf`.

   ```sh
   gcloud storage buckets create gs://lafeingcrokodil-flashcards-terraform \
     --project=lafeingcrokodil-flashcards \
     --location=<region> \
     --uniform-bucket-level-access \
     --public-access-prevention
   ```

   ```sh
   gcloud storage buckets update gs://lafeingcrokodil-flashcards-terraform --versioning
   ```

## Applying changes

1. Create a `terraform.tfvars` that defines the variables listed in
   `variables.tf`, optionally skipping any variables that already have
   defaults. (`project_name`, `billing_account_id` and `terraform_bucket` are
   no longer used, and can be deleted from an existing one.)
1. Initialize the directory.
    ```sh
    terraform init
    ```
1. Apply changes.
    ```sh
    terraform apply
    ```
