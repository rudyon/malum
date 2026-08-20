# Licensing

Malum's original source code is licensed as `AGPL-3.0-only`. The canonical
license text is the repository-root `LICENSE` file. This means version 3 is
the only version granted; Malum does not currently grant the optional
"any later version" permission.

The license applies to copyrightable work authored for Malum. It does not
replace the licenses of third-party libraries, icons, fonts, or tools. Their
current resolved inventory and distribution requirements are recorded in
`THIRD_PARTY_NOTICES.md`.

AGPLv3 is appropriate for a self-hosted server because a person who modifies
Malum and lets users interact with that modified version over a network must
offer those users the corresponding source. This requirement concerns the
program, not a user's imported documents, catalogue, or other library data.

## Release practice

The source repository contains dependency manifests, but not dependency source
trees or release artifacts. Before releasing a binary, container image, or
compiled frontend:

1. Run `pnpm run licenses` and inspect the resolved JavaScript dependency
   report.
2. Inspect the Go runtime graph with `go list -deps -f '{{if .Module}}{{.Module.Path}} {{.Module.Version}}{{end}}' ./...`.
3. Update `THIRD_PARTY_NOTICES.md` for dependency changes.
4. Package the applicable upstream copyright and license texts with the
   release, including the OFL information for bundled fonts.

This process is an engineering compliance practice, not legal advice. Obtain
legal advice before relying on it for a commercial or otherwise high-risk
distribution.
