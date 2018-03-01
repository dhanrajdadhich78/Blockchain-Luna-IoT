FROM golang:1.9

# if left blank app will run with dev settings

#ARG name
#ENV NODE_NAME $name
#ARG app_address
#ENV NODE_ADD $app_address
# it is okay to leave user/GoDoRP as long as you do not want to share code with other libraries
COPY . /go/src/wizeBlock/raft
WORKDIR /go/src/wizeBlock/raft

# added vendor services will need to be included here
#RUN go get ./vendor

RUN go get -v ./
RUN go build -v

# if dev setting will use pilu/fresh for code reloading via docker-compose volume sharing with local machine
# if production setting will build binary
#CMD ["./raft", "-id", "echo ${NODE_NAME}", "-haddr", "echo ${NODE_NAME}:11000", "-raddr", "echo ${NODE_NAME}:12000", "~/node"]
CMD ./raft ${NODE_COMMAND}

EXPOSE 11000 12000
