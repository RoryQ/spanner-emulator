FROM golang:1.22 AS builder
WORKDIR /build
COPY go.mod go.sum main.go ./
ENV CGO_ENABLED=0
RUN go build .

FROM gcr.io/cloud-spanner-emulator/emulator:1.5.28 AS runtime
COPY backend/query backend/query
COPY --from=builder /build/spanner-emulator ./
EXPOSE 9010 9020
CMD ["./spanner-emulator"]

