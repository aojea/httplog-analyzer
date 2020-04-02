FROM golang:1.13 AS builder

WORKDIR /go/src/httplog-analyzer
COPY . .
RUN go get -d -v ./...
RUN GOOS=linux CGO_ENABLED=0 go build -o /go/bin/httplog-analyzer

# STEP 2: Build Load Balancer Image
FROM scratch
COPY --from=builder /go/bin/httplog-analyzer /bin/httplog-analyzer
CMD ["/bin/httplog-analyzer"]
