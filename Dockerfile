FROM golang:latest
RUN go get github.com/rnurgaliyev/ipseek
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 \
    go build -o /ipseek \
    -v github.com/rnurgaliyev/ipseek

FROM arm32v7/alpine:latest
COPY --from=0 /ipseek /bin
EXPOSE 8088
CMD ["/bin/ipseek", "-c", "/etc/ipseek.yml"]
