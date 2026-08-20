# Third-party notices

This file records the third-party software and font licenses resolved for
Malum's source tree on 2026-08-20. It is an attribution inventory, not a
license change: each listed component remains under its own license.

`go.mod`, `go.sum`, `package.json`, and `pnpm-lock.yaml` are the authoritative
version locks. Before distributing a binary, container image, or compiled web
assets, regenerate the JavaScript report with `pnpm run licenses`, review the
Go module graph with `go list -deps -f '{{if .Module}}{{.Module.Path}} {{.Module.Version}}{{end}}' ./...`, and update this file and the distributed notices if either graph changes.

## Malum

Malum's original source code is `AGPL-3.0-only`; the full license text is in
[`LICENSE`](LICENSE).

## Go runtime dependencies

The Go application statically links the following resolved runtime modules.

| License | Modules |
| --- | --- |
| MIT | `codeberg.org/readeck/go-readability/v2` 2.1.2; `github.com/dustin/go-humanize` 1.0.1; `github.com/go-shiori/dom` 20230515143342-73569d674e1c; `github.com/gogs/chardet` 20211120154057-b7413eaefb8f; `github.com/itlightning/dateparse` 0.2.1; `github.com/mattn/go-isatty` 0.0.24; `github.com/ncruces/go-strftime` 1.0.0 |
| BSD-3-Clause | `github.com/andybalholm/cascadia` 1.3.4; `github.com/remyoudompheng/bigfft` 20230129092748-24d4a6f8daec; `golang.org/x/net` 0.58.0; `golang.org/x/sys` 0.47.0; `golang.org/x/text` 0.41.0; `modernc.org/libc` 1.75.3; `modernc.org/mathutil` 1.7.1; `modernc.org/memory` 1.12.0; `modernc.org/sqlite` 1.57.0 |

## Browser runtime dependencies and bundled fonts

| License | Components |
| --- | --- |
| MIT | `@dicebear/core` 10.6.0; `react` 19.2.8; `react-dom` 19.2.8; `react-router` and `react-router-dom` 7.18.2 |
| ISC | `lucide-react` 1.31.0 |
| CC0-1.0 OR MIT | `@dicebear/styles` 10.5.0 |
| OFL-1.1 | `@fontsource/atkinson-hyperlegible` 5.3.0; `@fontsource/cabin` 5.3.0; `@fontsource/merriweather` 5.3.0 |

The three font packages contain font files delivered to readers by the web
application. Their OFL copyright and license information must accompany any
distribution that includes those files.

## Build dependencies

The following tools are used to type-check or build Malum; they are not
currently distributed as part of the application.

| License | Components |
| --- | --- |
| MIT | `@tailwindcss/vite` 4.3.3; `@types/react` 19.2.18; `@types/react-dom` 19.2.4; `@vitejs/plugin-react` 6.0.5; `daisyui` 5.7.18; `tailwindcss` 4.3.3; `vite` 8.2.1, and their MIT-licensed transitive build dependencies |
| Apache-2.0 | `typescript` 5.9.3; `detect-libc` 2.1.2 |
| MPL-2.0 | `lightningcss` 1.32.0 and 1.33.0, including their Windows native bindings |
| BSD-3-Clause | `source-map-js` 1.2.1 |
| ISC | `graceful-fs` 4.2.11; `picocolors` 1.1.1 |

## Distribution responsibilities

Permissive licenses such as MIT, ISC, and BSD require their relevant copyright
and license notices to be retained when their code is distributed. For every
release artifact that contains third-party code or fonts, include the relevant
upstream license texts and copyright notices alongside this inventory. Do not
relicense those components as AGPL.

The current repository does not commit `node_modules`, Go module sources, a
compiled frontend, a binary, or a container image. Once Malum ships any of
those artifacts, its release process must package the applicable notices rather
than relying only on this source-tree inventory.
