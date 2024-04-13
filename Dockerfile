FROM alpine:latest

WORKDIR /root/

COPY log-server .

CMD ["./log-server"]