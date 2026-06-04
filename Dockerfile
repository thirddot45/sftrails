FROM golang:1-alpine AS build

RUN apk add --no-cache ca-certificates
# Pin the templ generator to the runtime version in go.mod. Using @latest can
# pull a generator that emits code referencing newer runtime APIs (e.g.
# templ.ResolveAttributeValue), which fails to compile against the pinned
# github.com/a-h/templ runtime. Keep this in sync with go.mod.
RUN go install github.com/a-h/templ/cmd/templ@v0.3.977

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN templ generate
RUN CGO_ENABLED=0 go build -tags postgres -ldflags="-s -w" -o /sftrails .

FROM scratch

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /sftrails /sftrails
COPY static/ /static/

EXPOSE 8080
ENTRYPOINT ["/sftrails"]
