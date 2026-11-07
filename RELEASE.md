# Releasing the Provider

## Prerequisites (one-time)

1. **GPG signing key.** See `~/.config/mise/gpg-yubikey-setup.md`. You need:

   - `gpg --armor --export-secret-subkeys <fingerprint>` → `GPG_PRIVATE_KEY`
   - the key passphrase → `GPG_PASS`
   - `gpg --armor --export <fingerprint>` → public key, for the registries

2. **Secrets** — Bitwarden Secrets Manager is the single source; `fnox push`
   syncs them into the repository's GitHub Actions secrets. Once per machine:

   ```sh
   gh auth login
   ```

   Create the remaining secrets in the bws project (key names matter):

   - `forward-email-api-token` — exists already
   - `gpg-private-key` — `gpg --armor --export-secret-subkeys <fingerprint>`
   - `gpg-pass` — the key passphrase

   Then sync:

   ```sh
   fnox push
   ```

   The mapping lives in the git-ignored `fnox.local.toml`; contributors never
   need it (CI secrets and their own local setups are independent).

3. **Register with the registries:**

   - **Terraform Registry:** sign in at <https://registry.terraform.io> with the
     GitHub account that owns this repo → Publish → Provider → select
     `smokeythemonkey/terraform-provider-forwardemail` → upload the GPG public key.
   - **OpenTofu Registry:** file an issue on
     [`opentofu/registry`](https://github.com/opentofu/registry) using the
     *Provider* issue template, and a second issue using the
     *Submit new Provider Signing Key* template with the ASCII-armored public key.

## Cutting a release

Tags must be `v`-prefixed semantic versions (e.g. `v0.1.0`), and no branch may
share the tag's name.

**CI (default):**

```sh
git tag v0.1.0
git push origin v0.1.0
```

The workflow builds, signs, and creates the GitHub release. Both registries pick
it up automatically.

**Local:**

```sh
git push origin main
git tag v0.1.0
read -s GITHUB_TOKEN; export GITHUB_TOKEN          # PAT with contents:write
export GPG_FINGERPRINT=$(gpg --list-secret-keys --with-colons | awk -F: '/^fpr/ {print $10; exit}')
mise run release
git push origin v0.1.0
```

**Checks:**

- `mise run release-check` — validate `.goreleaser.yaml`
- `mise run release-dry` — build all artifacts without publishing or signing

## Verify

```sh
terraform init   # or: tofu init
```

in a configuration using `source = "smokeythemonkey/forwardemail"`.
