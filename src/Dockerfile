FROM golang:1.27-trixie
ADD . /src
WORKDIR /src
ENV GOPATH=/install
RUN mkdir -p $GOPATH
RUN cd /src
RUN go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.2
RUN $GOPATH/bin/golangci-lint custom -v
RUN ./custom-gcl run
RUN go test ./...
RUN go install github.com/daixiang0/gci@latest
RUN go install golang.org/x/tools/cmd/goimports@latest
RUN go install github.com/go-delve/delve/cmd/dlv@latest
RUN go install honnef.co/go/tools/cmd/staticcheck@latest

FROM golang:1.27-trixie
COPY --from=0 /install/bin/* /bin/
COPY --from=0 /src/custom-gcl /bin/golangci-lint
RUN chmod +x /bin/*