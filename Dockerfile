FROM golang:alpine as builder
WORKDIR /opt/app
COPY . .

RUN go mod download
RUN go build -o /go/bin/app

FROM nginx:alpine
COPY --from=builder /go/bin/app /opt/app
COPY ./etc/default.conf /etc/nginx/conf.d/default.conf
RUN mkdir -p /usr/share/nginx/html/snapshots
RUN apk update && apk add --no-cache ffmpeg
COPY ./etc/start.sh /opt/start.sh
RUN ["chmod", "+x", "/opt/start.sh"]

ENV SNAPSHOT_DIRECTORY=/usr/share/nginx/html/snapshots
EXPOSE 80
ENTRYPOINT [ "/opt/start.sh" ]