# Contributing

1. [Fork the repository.][fork]
2. [Create a topic branch.][branch]
3. Implement your feature or bug fix.
4. Don't forget to add tests and make sure they pass (`go test -shuffle=on -count=1 -race ./...`).
5. `gofmt -s -w . && gofumpt -w . && go vet ./... && staticcheck ./... && golangci-lint run`.
6. `go run golang.org/x/tools/go/analysis/passes/modernize/cmd/modernize@latest -fix ./....`.
7. If necessary, add documentation for your feature or bug fix.
8. Commit and push your changes.
9. [Submit a pull request.][pr]

[fork]: http://help.github.com/fork-a-repo/
[branch]: https://help.github.com/en/github/collaborating-with-issues-and-pull-requests/about-branches
[pr]: https://help.github.com/en/github/collaborating-with-issues-and-pull-requests/about-pull-requests

## Release process

Kruk Shot uses GoReleaser and GitHub Actions to build and publish Windows releases. Executables are compressed with UPX and distributed as ZIP archives for AMD64 and ARM64. GitHub releases are immutable, so each release must use a new version tag.

1. Make sure all changes are committed and the project passes the checks:

   ```bash
   go test ./...
   go vet ./...
   go build ./...
   ```

2. Commit and push the changes to `master`:

   ```bash
   git add .
   git commit -m "Prepare v0.0.5 release"
   git push origin master
   ```

3. Create and push a new version tag:

   ```bash
   git tag -a v0.0.5 -m "Release v0.0.5"
   git push origin v0.0.5
   ```

4. GitHub Actions will automatically build and compress the Windows executables, generate ZIP archives and checksums, and publish the GitHub Release. GoReleaser generates the changelog from Git history.

**Do not create GitHub Releases manually.** Only push the Git tag and let GoReleaser handle publication. Published releases cannot be updated or replaced. If a release needs to be rebuilt after publication, use a new version tag.

View published releases [here](https://github.com/bodrovis-learning/kruk-shot/releases).