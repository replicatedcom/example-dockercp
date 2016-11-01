FROM golang:1.6.3

RUN go get -u github.com/kardianos/govendor # 7a0d30eab9e67b3ff60f13cdc483f003c01e04c1

# ENV GOPATH=$GOPATH:/dockercp
ENV PROJECTPATH=/go/src/github.com/divolgin/dockercp

WORKDIR $PROJECTPATH

CMD ["/bin/bash"]
