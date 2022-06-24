FROM golang:1.18-bullseye
RUN mkdir /app
ADD . /app
WORKDIR /app
RUN go get .
RUN go build -o main .
CMD ["/app/main"]