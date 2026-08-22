FROM alpine:latest

RUN apk add --no-cache git make musl-dev go

ENV GOROOT /usr/lib/go
ENV GOPATH /go
ENV PATH /go/bin:$PATH

RUN mkdir -p ${GOPATH}/src ${GOPATH}/bin

WORKDIR /app

COPY . .
RUN go mod tidy
RUN go mod download
RUN GOOS=linux GOARCH=amd64 go build -o app .

CMD ["/app/app"]
