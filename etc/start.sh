#!/bin/sh
# Start the application
nginx -g 'daemon off;' &
cd /opt/
exec /opt/app 
# start nginx
