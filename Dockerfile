FROM golang:1.15

RUN apt-get update && apt-get install -y git curl wait-for-it && rm -fr /var/lib/apt/lists/* && \
  go get -u -v github.com/go-task/task/cmd/task && \
  URL=https://github.com/golang-migrate/migrate/releases/download/v3.3.1/migrate.linux-amd64.tar.gz && \
  echo $URL && curl -#L $URL | tar -zxf - -C /go/bin/ && \
  mv -v /go/bin/migrate.linux-amd64 /go/bin/migrate


ADD ./ /go/src/github.com/treelightsoftware/pregxas-api
WORKDIR /go/src/github.com/treelightsoftware/pregxas-api

RUN task build