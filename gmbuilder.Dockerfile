FROM golang:1.20.3

WORKDIR /

ADD ./ /cosmos-sdk/

COPY --chmod=+x script.sh /script.sh
ENTRYPOINT /bin/bash /script.sh
