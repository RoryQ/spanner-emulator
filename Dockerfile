FROM golang:1.18 as builder
ENV GO111MODULE=on
WORKDIR /build
COPY go.mod go.sum main.go ./
RUN go build .

FROM gcr.io/cloud-spanner-emulator/emulator:1.5.0 as runtime
WORKDIR /gateway_main.runfiles/com_google_cloud_spanner_emulator/binaries/gateway_main_
COPY --from=builder /build/spanner-emulator ./
EXPOSE 9010 9020
CMD ["./spanner-emulator"]

