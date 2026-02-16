FROM golang:1-alpine AS build

RUN go install github.com/a-h/templ/cmd/templ@latest

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN templ generate
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /sftrails .

FROM scratch

COPY --from=build /sftrails /sftrails
COPY static/ /static/

EXPOSE 8080
ENTRYPOINT ["/sftrails"]
