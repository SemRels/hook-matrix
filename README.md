# hook-matrix

[![Latest Release](https://img.shields.io/github/v/release/SemRels/hook-matrix?label=version\&color=blue)](https://github.com/SemRels/hook-matrix/releases/latest)

Posts a release notification message to a Matrix room.

This plugin is distributed as the standalone Go binary `semrel-plugin-hook-matrix`. Semrel executes the binary as a subprocess, provides plugin configuration through `SEMREL_PLUGIN_*` environment variables, provides release context through `SEMREL_*` environment variables, reads standard output, and treats exit code `0` as success and any non-zero exit code as failure. Install the binary in `~/.semrel/plugins/` or anywhere on your `$PATH`.

## Installation

### Binary

```bash
go install github.com/SemRels/hook-matrix/cmd/plugin@latest
```

### Docker

Pre-built, multi-platform images (linux/amd64, linux/arm64) are published to the GitHub Container Registry on every release:

```bash
docker pull ghcr.io/semrels/hook-matrix:latest
```

Images are signed with [cosign](https://github.com/sigstore/cosign) and include a full SBOM attestation. Verify the signature:

```bash
cosign verify ghcr.io/semrels/hook-matrix:latest \
  --certificate-identity-regexp 'https://github.com/SemRels/hook-matrix/.github/workflows/release.yml.*' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com
```


## Configuration

```yaml
plugins:
  - name: hook-matrix
    path: ~/.semrel/plugins/semrel-plugin-hook-matrix
    env:
      SEMREL_PLUGIN_HOMESERVER_URL: "https://matrix.org"
      SEMREL_PLUGIN_TOKEN: "${MATRIX_TOKEN}"
      SEMREL_PLUGIN_ROOM_ID: "!abcdef:matrix.org"
      SEMREL_PLUGIN_MAX_RETRIES: "3"
      SEMREL_PLUGIN_RETRY_DELAY: "2s"
```

## `SEMREL_PLUGIN_*` variables

| Name | Required | Description | Default |
| --- | --- | --- | --- |
| `SEMREL_PLUGIN_HOMESERVER_URL` | Required | Matrix homeserver base URL. | None |
| `SEMREL_PLUGIN_TOKEN` | Required | Matrix access token. | None |
| `SEMREL_PLUGIN_ROOM_ID` | Required | Matrix room ID to post into. | None |
| `SEMREL_PLUGIN_MAX_RETRIES` | Optional | Retries on transient network failures and HTTP 5xx responses. | `3` |
| `SEMREL_PLUGIN_RETRY_DELAY` | Optional | Delay between retry attempts. | `2s` |

## Retry behavior

The plugin retries transient failures caused by network errors or HTTP `5xx` responses. HTTP `2xx` and `4xx` responses are not retried. Each retry attempt is logged to standard error.

## `SEMREL_*` release context used

| Variable | Description |
| --- | --- |
| `SEMREL_VERSION` | Resolved release version for the current run. |
| `SEMREL_TAG_NAME` | Git tag name semrel will create or publish. |
| `SEMREL_NEXT_VERSION` | Next version computed by semrel for the release. |
| `SEMREL_CHANGELOG` | Generated changelog text for the release. |
| `SEMREL_DRY_RUN` | Whether semrel is running in dry-run mode. |

## Example behavior

The plugin renders a Matrix message containing the release details and posts it to the configured room. In dry-run mode it prints the rendered message instead of sending it.

## License

Apache-2.0
