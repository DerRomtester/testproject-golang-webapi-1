FROM golang:1.22.2-bookworm
WORKDIR /app

COPY go.mod ./
RUN go mod download
RUN go mod verify

# copy directory files i.e all files ending with .go
COPY . /app
COPY conf/config.yaml /app
RUN go build -o devices-api ./cmd**

# tells Docker that the container listens on specified network ports at runtime
EXPOSE 8080

# command to be used to execute when the image is used to start a container
CMD [ "./devices-api" ]