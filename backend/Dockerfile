# Build context should be the repo root

ARG GO_VERSION=1.27

FROM golang:${GO_VERSION} AS build
WORKDIR /src

# Manifests first so dependency download caches independently of source edits.
COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./

# Placeholder: lexcheck is the only command that exists. Point COMMAND at the
# server once there is one.
ARG COMMAND=./cmd/lexcheck
RUN CGO_ENABLED=0 go build -trimpath -o /out/protocast ${COMMAND}

FROM gcr.io/distroless/static-debian12 AS runtime
WORKDIR /app
COPY --from=build /out/protocast /app/protocast
COPY lexicons/ /app/lexicons/

ENTRYPOINT ["/app/protocast"]
CMD ["/app/lexicons"]
