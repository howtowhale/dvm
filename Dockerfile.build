FROM golang:1.5

ADD . $GOPATH/src/github.com/howtowhale/dvm/
WORKDIR $GOPATH/src/github.com/howtowhale/dvm/

RUN make get-deps
