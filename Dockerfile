FROM golang

ADD . /go/src/github.com/CharlesMassry/meme_generator
RUN go get github.com/golang/freetype
RUN go get github.com/golang/freetype/truetype
RUN go get github.com/satori/go.uuid

WORKDIR /go/src/github.com/CharlesMassry/meme_generator
RUN go build -o main

CMD ["./main", "-PORT", "`$PORT`"]
