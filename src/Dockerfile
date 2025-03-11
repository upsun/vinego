FROM golang:1.24-bookworm
ADD . /src
WORKDIR /src
ENV GOPATH=/install
RUN mkdir -p $GOPATH
RUN cd /src
RUN go test ./...
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint
RUN $GOPATH/bin/golangci-lint custom -v
RUN $GOPATH/bin/golangci-lint run
RUN go install github.com/daixiang0/gci
RUN go install golang.org/x/tools/cmd/goimports
RUN go install github.com/go-delve/delve/cmd/dlv
RUN go install honnef.co/go/tools/cmd/staticcheck

FROM golang:1.24-bookworm
COPY --from=0 /install/bin/* /bin/
RUN chmod +x /bin/*
COPY --from=0  src/vinego.so /custom_linters/
