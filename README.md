# hook-matrix

`hook-matrix` is a SemRel hook plugin that posts release announcements to a Matrix room using the Matrix client HTTP API.

## Configuration

Environment variables:

- `MATRIX_HOMESERVER`
- `MATRIX_ROOM_ID`
- `MATRIX_ACCESS_TOKEN`

## Behavior

The hook sends a `m.room.message` event containing the release version, repository, and changelog.

## Development

```bash
go mod tidy
go build ./...
go test ./...
```
