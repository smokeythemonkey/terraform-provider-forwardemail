# Contributing

## Development setup

- [mise](https://mise.jdx.dev/) provides the entire toolchain (Go, linters, terraform,
  goreleaser, fnox) — run `mise install` after cloning.
- `mise run build`, `mise run lint`, `mise run docs` build, lint, and regenerate docs.

## Running the tests

Acceptance tests hit the real Forward Email API and need an API key
(`FORWARDEMAIL_API_KEY`). Either export it in your shell:

```sh
FORWARDEMAIL_API_KEY=... mise run test
```

or create a personal `fnox.local.toml` (git-ignored) that pulls it from your own
secret manager. For example, with Bitwarden Secrets Manager:

```toml
[providers.bws]
type = "bitwarden-sm"
project_id = "<your project id>"

[secrets]
FORWARDEMAIL_API_KEY = { provider = "bws", value = "<secret key name>" }
```

Any [fnox provider](https://fnox.jdx.dev/providers/overview) works (age, 1Password,
OS keychain, ...). Without a key the tests skip.

`mise run test` runs the acceptance suite and then the resource sweepers; the
pre-push hook runs the same command. Pull requests additionally run the suite in
CI against several Terraform versions.

Please defer to the [GitHub Contributing document](https://github.com/github/.github/blob/master/CONTRIBUTING.md)
for everything else.
