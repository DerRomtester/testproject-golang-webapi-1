FROM golang:1.22.2-bookworm
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
RUN go mod verify

# copy directory files i.e all files ending with .go
COPY . /app
COPY conf/config.yaml /app
RUN go test -v ./...
RUN CGO_ENABLED=0 GOOS=linux go build -o devices-api ./cmd**

EXPOSE 8080

# command to be used to execute when the image is used to start a container
CMD [ "./devices-api" ]