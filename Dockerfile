FROM golang:1.18-bullseye
ENV app /app
RUN mkdir -p $app
WORKDIR $app
ADD . $app
RUN go build -o main
RUN rm -rf *go *.mod *.sum bin services utils
CMD ./main


