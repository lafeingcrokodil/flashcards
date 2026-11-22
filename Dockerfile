# Base images are pinned by digest as well as tag, so that builds are
# reproducible. The Node and Go versions must match web/.nvmrc and go.mod,
# which CI checks.

# Build the front end.
FROM node:26.13.0-trixie-slim@sha256:5adad64f9a1fca2d51972caa65983634c8a774eb300d72b20f7f2a0f60368072 AS web
WORKDIR /src/web
# The manifests come before the sources, so that the installed dependencies
# are cached while only the sources change.
COPY web/package.json web/pnpm-lock.yaml web/pnpm-workspace.yaml web/.npmrc ./
# Install the pnpm version from the packageManager field in package.json.
RUN npm install --global "$(node --print "require('./package.json').packageManager")"
RUN pnpm install --frozen-lockfile
COPY web/ ./
RUN pnpm build

# Build the server, with the front end embedded.
FROM golang:1.27.4-trixie@sha256:f18c3d73a184fb0582828510722e723b63a93f7188d11dd223d9be6f6880e65d AS server
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY internal/ internal/
COPY web/web.go web/
COPY --from=web /src/web/build web/build
# A static binary, so that it runs on an image without a C library.
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /flashcards ./cmd/flashcards

# The final image contains only the binary: no shell or package manager. The
# nonroot variant runs as an unprivileged user.
FROM gcr.io/distroless/static-debian13:nonroot@sha256:aafe993665bf6200db38a02c11a9351e782a84b6a45e7f33b6695638c88e3931
COPY --from=server /flashcards /flashcards
EXPOSE 8080
ENTRYPOINT ["/flashcards"]
