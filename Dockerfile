FROM golang:1.21 as builder
WORKDIR /build
COPY go.mod go.sum main.go ./
ENV CGO_ENABLED=0
RUN go build .

FROM gcr.io/cloud-spanner-emulator/emulator:1.5.17 as runtime
COPY backend/query backend/query
COPY --from=builder /build/spanner-emulator ./
EXPOSE 9010 9020
CMD ["./spanner-emulator"]

