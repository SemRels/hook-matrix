# hook-matrix

Posts a release notification message to a Matrix room.

This plugin is distributed as the standalone Go binary `semrel-plugin-hook-matrix`. Semrel executes the binary as a subprocess, provides plugin configuration through `SEMREL_PLUGIN_*` environment variables, provides release context through `SEMREL_*` environment variables, reads standard output, and treats exit code `0` as success and any non-zero exit code as failure. Install the binary in `~/.semrel/plugins/` or anywhere on your `$PATH`.

## Installation

```bash
go install github.com/SemRels/hook-matrix/cmd/plugin@latest
```

## Configuration

```yaml
plugins:
  - name: hook-matrix
    path: ~/.semrel/plugins/semrel-plugin-hook-matrix
    env:
      SEMREL_PLUGIN_HOMESERVER: "https://matrix.org"
      SEMREL_PLUGIN_TOKEN: "${MATRIX_TOKEN}"
      SEMREL_PLUGIN_ROOM_ID: "!abcdef:matrix.org"
      SEMREL_PLUGIN_MESSAGE_TEMPLATE: "Released {{ .TagName }}"
```

## `SEMREL_PLUGIN_*` variables

| Name | Required | Description | Default |
| --- | --- | --- | --- |
| `SEMREL_PLUGIN_HOMESERVER` | Required | Matrix homeserver base URL. | None |
| `SEMREL_PLUGIN_TOKEN` | Required | Matrix access token. | None |
| `SEMREL_PLUGIN_ROOM_ID` | Required | Matrix room ID to post into. | None |
| `SEMREL_PLUGIN_MESSAGE_TEMPLATE` | Optional | Template used to render the Matrix message body. | Built-in template |

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
