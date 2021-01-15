FROM alpine:latest

ARG REPO="loveandpeople-DAG/goBee-pkg"
ARG TAG=latest
ARG ARCH=x86_64
ARG OS=Linux

LABEL org.label-schema.description="LP-NODE - The LP community node"
LABEL org.label-schema.name="chainking/lp_node"
LABEL org.label-schema.schema-version="1.0"
LABEL org.label-schema.vcs-url="https://hub.docker.com/r/chainking/lp_node"

WORKDIR /app

RUN apk --no-cache add ca-certificates curl jq tini tar\
 && update-ca-certificates 2>/dev/null || true\
 && if [ "$TAG" = "latest" ];\
    then\
      TARBALL_URL=$(curl --retry 3 -f -s https://api.github.com/repos/${REPO}/releases/latest | jq -r .tarball_url);\
    else\
      TARBALL_URL=$(curl --retry 3 -f -s https://api.github.com/repos/${REPO}/releases/${TAG} | jq -r .tarball_url);\
    fi\
 && echo "Downloading from ${TARBALL_URL}"\
 && curl -f -L --retry 3 ${TARBALL_URL} -o /tmp/hornet.tgz\
 && tar --wildcards --strip-components=1 -xf /tmp/hornet.tgz -C /app/ */hornet */config.json */peering.json */snapshotMainnet.txt */plugins\
 && if [ "$ARCH" = "x86_64" ];\
    then\
      curl -f -L --retry 3 -o /etc/apk/keys/sgerrand.rsa.pub https://alpine-pkgs.sgerrand.com/sgerrand.rsa.pub;\
      curl -f -L --retry 3 -o glibc-2.30-r0.apk https://github.com/sgerrand/alpine-pkg-glibc/releases/download/2.30-r0/glibc-2.30-r0.apk;\
      apk add glibc-2.30-r0.apk;\
      rm glibc-2.30-r0.apk;\
    fi\
 && addgroup --gid 39999 hornet\
 && adduser -h /app -s /bin/sh -G hornet -u 39999 -D hornet\
 && chmod +x /app/hornet\
 && chown hornet:hornet -R /app\
 && rm /tmp/hornet.tgz\
 && apk del jq curl tar

USER hornet
ENTRYPOINT ["/sbin/tini", "--", "/app/hornet"]

