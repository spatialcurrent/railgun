FROM gcr.io/distroless/base
MAINTAINER Spatial Current, Inc.

ADD bin/railgun_linux_amd64 /railgun

EXPOSE 8080

CMD ["/railgun"]