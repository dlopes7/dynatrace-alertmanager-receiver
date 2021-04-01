FROM golang:1.15

WORKDIR /tmp/dynatrace-receiver

COPY . .
RUN go build -o /out/dynatrace-receiver .

CMD ["/out/dynatrace-receiver"]