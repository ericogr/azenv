ARG GO_VERSION
FROM golang:${GO_VERSION}-alpine AS builder
RUN apk add --no-cache ca-certificates make git curl gcc libc-dev
RUN mkdir -p /build
WORKDIR /build
COPY go.mod /build
COPY go.sum /build
COPY main.go /build
COPY cmd /build/cmd
COPY services /build/services
COPY Makefile /build
COPY .git /build/.git
RUN go mod download
RUN make build-linux

FROM golang:${GO_VERSION}-alpine 
RUN apk add --no-cache ca-certificates bash git gcc libc-dev openssh
ENV GO111MODULE on
COPY --from=builder /build/azenv /bin/azenv
ENTRYPOINT ["bash"]