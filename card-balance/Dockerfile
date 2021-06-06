FROM golang:1.16 as builder

WORKDIR /go/src/app
COPY . .

RUN echo "=> start get dependent" \
    && go mod tidy
    
RUN echo "=> start compile" \
    && CGO_ENABLED=0 go build -o main .\
    && echo "=> compile complete" 


FROM alpine:3.13
WORKDIR /root/
COPY --from=builder /go/src/app/main .
CMD [ "./main" ]
