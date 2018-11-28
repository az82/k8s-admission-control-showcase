FROM alpine:3.8

EXPOSE 8080
EXPOSE 8443

VOLUME ["certs"]

ADD main-linux-amd64 /
CMD [ "/main-linux-amd64" , "-stderrthreshold=INFO"]