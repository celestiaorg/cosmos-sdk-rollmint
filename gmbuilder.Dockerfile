FROM golang:1.20.3

WORKDIR /

ADD ./ /cosmos-sdk/

COPY script.sh /script.sh
RUN chmod +x /script.sh
ENTRYPOINT /bin/bash /script.sh
