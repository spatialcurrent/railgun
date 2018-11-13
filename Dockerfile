FROM gcr.io/distroless/base
MAINTAINER Spatial Current, Inc.

ADD bin/chamber_linux_amd64 /chamber
ADD bin/railgun_linux_amd64 /railgun

EXPOSE 8080

ENTRYPOINT ["/bin/chamber", "exec", "railgun-prod", "--", "/bin/railgun"]

CMD ["--verbose"]