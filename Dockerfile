FROM golang:1.12 as build

ENV GOOS linux
ENV GOARCH amd64
ENV CGO_ENABLED 0
ENV GO111MODULE on

COPY . /go/src/github.com/lamoda/gonkey
WORKDIR /go/src/github.com/lamoda/gonkey
RUN make build

FROM alpine:3.9
LABEL Author="Denis Sheshnev <denis.sheshnev@lamoda.ru>"

COPY --from=build /go/src/github.com/lamoda/gonkey/gonkey /bin/gonkey
ENTRYPOINT ["/bin/gonkey"]
CMD ["-spec=/gonkey/swagger.yaml", "-host=${HOST_ARG}"]
