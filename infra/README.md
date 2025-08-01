# Infrastructure

1. Create a `terraform.tfvars` that defines the variables listed in `variables.tf`, optionally skipping any variables that already have defaults.
1. Initialize the directory, replacing `<terraform_bucket>` with whatever the `terraform_bucket` variable is set to in `terraform.tfvars`.
    ```sh
    terraform init -backend-config="bucket=<terraform_bucket>"
    ```
1. Apply changes.
    ```sh
    terraform apply
    ```
