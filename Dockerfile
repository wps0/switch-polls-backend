FROM golang:1.17

WORKDIR src
COPY . .

RUN cd /go/src/src/ && go get -d -v ./
RUN cd /go/src/src/ && go install -v ./

CMD ["/go/bin/switch-polls-backend"]