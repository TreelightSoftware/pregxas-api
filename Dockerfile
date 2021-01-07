FROM golang:1.15

RUN apt-get install -y git curl && \
  go get -u -v github.com/go-task/task/cmd/task && \
  go get github.com/t-yuki/gocover-cobertura && \
  mkdir -p /go/src/github.com/treelightsoftware && \
  mkdir -p /root/.ssh && \
  echo "IdentityFile /root/.ssh/id_rsa" >> /etc/ssh/ssh_config && \
  echo "Host github.com\n    StrictHostKeyChecking no\n" >> /root/.ssh/config \
  echo "Host bitbucket.org\n    StrictHostKeyChecking no\n" >> /root/.ssh/config

ADD ./ /go/src/github.com/treelightsoftware/pregxas-api
WORKDIR /go/src/github.com/treelightsoftware/pregxas-api

RUN task build
CMD ["task", "run"]