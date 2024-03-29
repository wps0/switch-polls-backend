FROM golang:1.18.2

WORKDIR src
COPY . .

RUN cd /go/src/src/ && go get -d -v ./
RUN cd /go/src/src/ && go install -v ./

CMD ["/go/bin/switch-polls-backend","-cfg","/go/src/cfg/config.json"]
