FROM golang:1.24-bookworm
ADD . /src
WORKDIR /src
ENV GOPATH=/install
RUN mkdir -p $GOPATH
RUN cd /src
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.64.6
RUN $GOPATH/bin/golangci-lint custom -v
RUN ./custom-gcl run
RUN go test ./...
RUN go install github.com/daixiang0/gci@latest
RUN go install golang.org/x/tools/cmd/goimports@latest
RUN go install github.com/go-delve/delve/cmd/dlv@latest
RUN go install honnef.co/go/tools/cmd/staticcheck@latest

FROM golang:1.24-bookworm
COPY --from=0 /install/bin/* /bin/
COPY --from=0 /src/custom-gcl /bin/golangci-lint
RUN chmod +x /bin/*