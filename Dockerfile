FROM golang:1.20
WORKDIR /usr/src/app
COPY src/go.mod src/go.sum ./
RUN go mod download && go mod verify

COPY src/ .
RUN go build -o /usr/local/bin/cnetmon .
CMD ["cnetmon"]