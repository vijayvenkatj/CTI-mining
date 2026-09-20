FROM golang:1.26-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /out/ingestion ./cmd/ingestion
RUN CGO_ENABLED=0 go build -o /out/edge-generation ./cmd/edge-generation
RUN CGO_ENABLED=0 go build -o /out/estimator ./cmd/estimator
RUN CGO_ENABLED=0 go build -o /out/server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /out/ ./
