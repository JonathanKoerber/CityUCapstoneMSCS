FROM golang:1.24-bookworm AS base

WORKDIR /plc-node

COPY go.mod go.sum ./

RUN go mod download

COPY ./  .

EXPOSE 502

RUN go build -o modbusNode /plc-node/main.go

CMD ["/plc-node/modbusNode"]