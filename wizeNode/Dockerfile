FROM golang:1.9

RUN apt-get update && apt-get install libgmp-dev -y

# if left blank app will run with dev settings
# to build production image run:
# $ docker build ./nodeWize --build-args app_env=production
#ARG app_port
#ENV NODE_ID $app_port
#ARG app_address
#ENV NODE_ADD $app_address
# it is okay to leave user/GoDoRP as long as you do not want to share code with other libraries
COPY . /go/src/wizeBlock/wizeNode
WORKDIR /go/src/wizeBlock/wizeNode

ENV CGO_LDFLAGS_ALLOW=".*\.a"

# added vendor services will need to be included here
#RUN go get ./vendor

RUN go get -v ./
RUN go build -v

# if dev setting will use pilu/fresh for code reloading via docker-compose volume sharing with local machine
# if production setting will build binary
#CMD ["./wizeNode", "startnode"]
CMD ./wizeNode ${NODE_COMMAND}

EXPOSE 3000 4000
