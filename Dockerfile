FROM golang:1.25 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /gitbackup .

FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends git ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && groupadd -g 65532 nonroot \
    && useradd -r -u 65532 -g nonroot -d /home/nonroot -m nonroot

COPY --from=builder /gitbackup /gitbackup

USER nonroot:nonroot

ENTRYPOINT ["/gitbackup"]
