## Overview

This guide sets up a local Jenkins LTS instance in Docker so you can run this repo's `Jenkinsfile` before using a shared Jenkins server. By the end, Jenkins will be running at `http://localhost:8080`, the required plugins and toolchain will be installed, and a `craftsbite-deploy` pipeline job will be ready to execute this project's Terraform-based pipeline.

## Prerequisites

- Run every command from the repo root, the directory that contains `Jenkinsfile`.
- Docker Desktop or Docker Engine must be installed and running. `local-testing.md` does not pin a Docker version; the image used here is `jenkins/jenkins:lts`.
- Git must be installed on the host. `local-testing.md` does not pin a Git version.
- A browser must be available to open `http://localhost:8080`.
- The Jenkins container needs outbound internet access to download Jenkins plugins, Debian packages, Terraform, Go, and AWS CLI.
- You need an AWS access key ID and secret access key that can access the Terraform backend bucket `trainee-2026-sayad-craftsbite`, operate in `ap-southeast-1`, and update the Lambda artifacts managed by this repo's Terraform code.
- Jenkins will need these plugins during setup: the setup wizard's suggested plugins, `workflow-aggregator`, `git`, `credentials-binding`, `workflow-cps`, and `aws-credentials`.
- The toolchain installed into the container must satisfy these versions: Terraform `>= 1.5` from `terraform/versions.tf`, Go `1.25.0` from `local-testing.md`, AWS CLI v2, and `zip`/`unzip`.
- macOS and Linux can use the Bash commands below as written.
- Windows PowerShell needs the PowerShell `docker run` command in Step 1. The in-container commands stay the same because the Jenkins container is Linux-based.

## Setup

1. Start Jenkins with the repo bind-mounted from the beginning.

   **Deviation from `.idea/jenkins/local-testing.md`:** the source note starts Jenkins once without a repo mount and later recreates the container with one. This guide uses the bind-mounted `docker run` command up front so the pipeline job can see the repo immediately.

   macOS or Linux:

   ```bash
   docker run -d \
     --name jenkins-local \
     -p 8080:8080 \
     -p 50000:50000 \
     -v jenkins_home:/var/jenkins_home \
     -v "$(pwd):/var/jenkins_home/workspace/craftsbite-deploy" \
     jenkins/jenkins:lts
   ```

   Windows PowerShell:

   ```powershell
   docker run -d `
     --name jenkins-local `
     -p 8080:8080 `
     -p 50000:50000 `
     -v jenkins_home:/var/jenkins_home `
     -v "${PWD}:/var/jenkins_home/workspace/craftsbite-deploy" `
     jenkins/jenkins:lts
   ```

   **Run this from the repo root.** If you launch the container from another directory, Jenkins will mount the wrong host path and the pipeline job will not be able to read this repo.

2. Wait about 30 seconds, then retrieve the initial Jenkins unlock password.

   ```bash
   docker exec jenkins-local cat /var/jenkins_home/secrets/initialAdminPassword
   ```

   Open `http://localhost:8080`. Jenkins should show the unlock screen. Paste the password from the command output to continue.

3. Finish the first-run wizard, then install the repo-specific plugins.

   In the browser, click `Install suggested plugins`. Wait for the setup wizard to finish, then create an admin user or continue with the generated account.

   After the setup wizard completes, run:

   ```bash
   docker exec jenkins-local bash -c '
     jenkins-plugin-cli --plugins \
       workflow-aggregator \
       git \
       credentials-binding \
       workflow-cps \
       aws-credentials
   '
   ```

   Restart Jenkins so the plugins are active:

   ```bash
   docker restart jenkins-local
   ```

   **Deviation from `.idea/jenkins/local-testing.md`:** the current `Jenkinsfile` uses `AmazonWebServicesCredentialsBinding`, so `aws-credentials` is required even though it is not listed in the older note.

4. Install the required build and deployment tools inside the container.

   **Deviation from `.idea/jenkins/local-testing.md`:** the original command assumes `curl` and `wget` are already available in the image. The command below installs them explicitly so the step works on a fresh `jenkins/jenkins:lts` container.

   ```bash
   docker exec -u root jenkins-local bash -c '

     # --- zip + download tools ---
     apt-get update -qq && apt-get install -y -qq zip unzip curl wget > /dev/null

     # --- Terraform ---
     apt-get install -y -qq gnupg software-properties-common > /dev/null
     wget -qO- https://apt.releases.hashicorp.com/gpg | gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg
     echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(. /etc/os-release && echo $VERSION_CODENAME) main" > /etc/apt/sources.list.d/hashicorp.list
     apt-get update -qq && apt-get install -y -qq terraform > /dev/null

     # --- Go ---
     GO_VERSION=1.25.0
     wget -q https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz
     tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz
     rm go${GO_VERSION}.linux-amd64.tar.gz
     ln -s /usr/local/go/bin/go /usr/local/bin/go

     # --- AWS CLI v2 ---
     curl -s "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o /tmp/awscliv2.zip
     unzip -q /tmp/awscliv2.zip -d /tmp
     /tmp/aws/install
     rm -rf /tmp/aws /tmp/awscliv2.zip

   '
   ```

   Verify the tools are installed:

   ```bash
   docker exec jenkins-local bash -c '
     terraform version
     go version
     zip -h | head -1
     aws --version
   '
   ```

5. Configure AWS credentials in Jenkins.

   Open `http://localhost:8080`, then go to `Manage Jenkins` > `Credentials` > `System` > `Global credentials` > `Add Credentials`.

   Set these fields:

   - `Kind`: `AWS Credentials`
   - `Scope`: `Global`
   - `ID`: `aws-credentials`
   - `Access Key ID`: your AWS access key ID
   - `Secret Access Key`: your AWS secret access key

   Save the credential.

   **Deviation from `.idea/jenkins/local-testing.md`:** the older note says to create two `Secret text` credentials with IDs `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`. The current `Jenkinsfile` does not read those IDs. It expects one AWS credential with ID `aws-credentials`.

6. Create the pipeline job.

   From the Jenkins dashboard, click `New Item`. Enter `craftsbite-deploy`, select `Pipeline`, and click `OK`.

   In the `Pipeline` section, set:

   - `Definition`: `Pipeline script from SCM`
   - `SCM`: `Git`
   - `Repository URL`: `file:///var/jenkins_home/workspace/craftsbite-deploy`
   - `Branch Specifier`: `*/main`
   - `Script Path`: `Jenkinsfile`

   Click `Save`.

   **Current pipeline behavior:** `local-testing.md` still describes `main` as the branch that reaches `Apply`, but the checked-in `Jenkinsfile` now uses `expression { env.GIT_BRANCH?.endsWith('/jenkins') }`. A `main` build will still run `Init`, `Validate`, and `Plan`, but it will skip `Apply`.

7. Run the pipeline and inspect the console output.

   Open `http://localhost:8080/job/craftsbite-deploy/`, click `Build Now`, click the new build number in `Build History`, then click `Console Output`.

   The pipeline defines these stages in order: `Init`, `Validate`, `Plan`, and `Apply`. With the `*/main` branch specifier from Step 6, expect `Apply` to be skipped because the checked-in `when` condition does not match `main`.

   Look for these successful outputs:

   - `Init`: `Terraform has been successfully initialized!`
   - `Validate`: `Success! The configuration is valid.`
   - `Plan`: `Plan: X to add, 0 to change, 0 to destroy.`
   - `Apply` on a default `*/main` run: `Stage "Apply" skipped due to when conditional`
   - `Apply` when the branch condition matches or is temporarily forced: `Apply complete! Resources: X added, 0 changed, 0 destroyed.`

   If you only need to test `Apply` locally, temporarily change the checked-in condition:

   ```groovy
   when {
       expression { env.GIT_BRANCH?.endsWith('/jenkins') }
   }
   ```

   to:

   ```groovy
   when {
       expression { true }
   }
   ```

   **Revert that change before committing.**

## Verification

- Open `http://localhost:8080`. Before setup is complete, Jenkins should render the unlock screen. After setup is complete, it should render the Jenkins dashboard.
- Jenkins runs on port `8080`. The container also exposes port `50000` for Jenkins inbound agents.
- This command should show the container as running:

  ```bash
  docker ps --filter name=jenkins-local
  ```

- This command should eventually include `Jenkins is fully up and running`:

  ```bash
  docker logs jenkins-local
  ```

- This command returns the one-time unlock password used on the first screen:

  ```bash
  docker exec jenkins-local cat /var/jenkins_home/secrets/initialAdminPassword
  ```

- A successful default run from this guide shows `Init`, `Validate`, and `Plan` without errors, then `Stage "Apply" skipped due to when conditional`. `Apply` succeeds only when the current branch condition matches or when you temporarily force it to `true`.

## Configuration reference

| Item | Type | Value | Purpose |
| --- | --- | --- | --- |
| `Jenkinsfile` | config file | repo root | Declares the Jenkins pipeline Jenkins executes. |
| `terraform/versions.tf` | config file | repo path | Sets Terraform `required_version`, the S3 backend bucket/key, and the AWS provider version. |
| `terraform/main.tf` | config file | repo path | Builds Lambda zip files, uploads them, and updates Lambda functions through local `bash` and AWS CLI calls. |
| `8080` | port | host -> Jenkins UI | Serves the Jenkins web interface. |
| `50000` | port | host -> Jenkins inbound agents | Exposes the default Jenkins agent port. |
| `jenkins_home` | Docker volume | named volume | Persists Jenkins state, plugins, jobs, and credentials across container restarts. |
| `/var/jenkins_home/workspace/craftsbite-deploy` | bind mount path | container path | Makes the checked-out repo visible to Jenkins inside the container. |
| `aws-credentials` | Jenkins credential | credential ID | The credential ID the current `Jenkinsfile` reads through `AWS_CREDENTIALS_ID`. |
| `AWS_DEFAULT_REGION` | environment variable | `ap-southeast-1` | Region exported by the current `Jenkinsfile` for Terraform and AWS CLI steps. |
| `AWS_CREDENTIALS_ID` | environment variable | `aws-credentials` | Tells the current `Jenkinsfile` which Jenkins credential to bind. |
| `AWS_ACCESS_KEY_ID` | environment variable | injected at runtime | Filled by `withCredentials` inside Jenkins shell steps. |
| `AWS_SECRET_ACCESS_KEY` | environment variable | injected at runtime | Filled by `withCredentials` inside Jenkins shell steps. |
| `trainee-2026-sayad-craftsbite` | Terraform backend setting | S3 bucket name | Stores Terraform state for this project. |
| `terraform/state/terraform.tfstate` | Terraform backend setting | S3 object key | Stores the remote Terraform state file inside the backend bucket. |

## Troubleshooting

- **`AmazonWebServicesCredentialsBinding` is missing or unknown.** The `aws-credentials` plugin is not installed. Run the Step 3 plugin install command, then restart Jenkins.
- **The build says it cannot find credential `aws-credentials`.** The current `Jenkinsfile` expects one Jenkins credential with ID `aws-credentials`. Do not use the older two-secret `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` setup from `.idea/jenkins/local-testing.md`.
- **`terraform`, `go`, `zip`, or `aws` is missing when the pipeline runs.** Re-run Step 4. The `jenkins_home` volume persists Jenkins state, but the binaries installed into the container filesystem do not survive a fresh container.
- **`Apply` is skipped even though the rest of the pipeline succeeded.** This is expected when the Jenkins branch name does not end with `/jenkins`. The current `Jenkinsfile` no longer gates `Apply` on `main`.
- **`terraform init` fails with S3 or access-denied errors.** The AWS credential does not have access to backend bucket `trainee-2026-sayad-craftsbite` in `ap-southeast-1`, or it cannot update the Lambda resources used by this repo.
- **The job cannot read the repo or Jenkinsfile.** Start the container from the repo root, confirm the bind mount in Step 1, and keep the repository URL set to `file:///var/jenkins_home/workspace/craftsbite-deploy`.
- **Step 4 fails on `curl` or `wget`.** Use the Step 4 command from this README, not the older shorter variant. This README installs both tools explicitly before downloading Terraform, Go, and AWS CLI.

## Teardown

1. Stop the Jenkins container.

   ```bash
   docker stop jenkins-local
   ```

2. Remove the Jenkins container.

   ```bash
   docker rm jenkins-local
   ```

3. Remove the persistent Jenkins volume.

   ```bash
   docker volume rm jenkins_home
   ```

4. Verify that neither the container nor the volume still exists.

   ```bash
   docker ps -a --filter name=jenkins-local
   docker volume ls --filter name=jenkins_home
   ```

   Cleanup is complete when those commands show only their headers and no `jenkins-local` container or `jenkins_home` volume rows.
