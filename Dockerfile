FROM golang:alpine as builder
WORKDIR /opt/app
COPY . .

RUN go mod download
RUN go build -o /go/bin/app

FROM nginx:alpine
COPY --from=builder /go/bin/app /opt/app
COPY ./etc/nginx.conf /etc/nginx/nginx.conf
RUN mkdir -p /var/www/html/snapshots
RUN apk update && apk add --no-cache ffmpeg

ENV SNAPSHOT_DIRECTORY=/var/www/html/snapshots
EXPOSE 80
CMD ["/opt/app"] 
