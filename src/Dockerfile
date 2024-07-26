FROM golang:1.22-bullseye
ADD . /src
WORKDIR /src
ENV GOPATH=/install
RUN mkdir -p $GOPATH
RUN cd /src
RUN go build -buildmode=plugin vinego.go
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint
RUN go install github.com/daixiang0/gci
RUN go install golang.org/x/tools/cmd/goimports
RUN go install github.com/go-delve/delve/cmd/dlv
RUN go install honnef.co/go/tools/cmd/staticcheck

FROM golang:1.22-bullseye
COPY --from=0 /install/bin/* /bin/
RUN chmod +x /bin/*
COPY --from=0  src/vinego.so /custom_linters/
