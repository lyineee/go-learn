FROM golang:1.18 as builder

WORKDIR /go/src/app
COPY ["go.work", "main.go", "./"]

RUN echo "=> start get dependent" \
    && go mod tidy

RUN echo "=> start compile" \
    && CGO_ENABLED=0 go build -ldflags '-s -w'  -o main main.go \
    && echo "=> compile complete" 


FROM alpine:3.13
ENV TIME_ZONE=Asia/Shanghai
RUN apk update && \
    apk add --no-cache tzdata && \
    cp /usr/share/zoneinfo/$TIME_ZONE /etc/localtime && echo $TIME_ZONE > /etc/timezone && \
    apk del tzdata
WORKDIR /root/
COPY --from=builder /go/src/app/main .
CMD [ "./main" ]
