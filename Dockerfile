############################
# STEP 1 build executable binary
############################
FROM golang:1.12-stretch AS builder
COPY . /app/src/
ENV GOPROXY=https://mirrors.aliyun.com/goproxy/
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
#COPY $GOPATH /go
WORKDIR /app/src/

RUN go build -o /app/src/builds/server /app/src/main.go
#RUN go build -o /app/src/builds/artisan /app/src/artisan.go

############################
# STEP 2 build a small server image
############################
FROM scratch
# Copy .env.json
COPY --from=builder /app/src/.env.example.json /mindav/.env.json
# Copy our static executable.
COPY --from=builder /app/src/builds/server /mindav/server
#COPY --from=builder /app/src/builds/artisan /bin/artisan
WORKDIR /mindav/
# Run the server binary.
ENTRYPOINT ["/mindav/server"]
EXPOSE 80
