FROM  golang:1.26-alpine AS build

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -trimpath -o /bin/dfrecap .

RUN mkdir -p /empty-logs

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /

COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build /bin/dfrecap /devfest-recap

COPY conf.yaml conf.yaml

COPY frontend ./frontend

COPY --from=build --chown=nonroot:nonroot /empty-logs /logs

ENV TZ=Europe/Istanbul

EXPOSE 1923

USER nonroot:nonroot
ENTRYPOINT [ "/devfest-recap" ]