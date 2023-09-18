FROM golang:1.19 as builder

WORKDIR /app

COPY . .

RUN go build -o podassrt


FROM alpine:3.9

LABEL authors="eddielth"

WORKDIR /app

COPY --from=builder /app/podassrt /app/podassrt
COPY --from=builder /app/res/configuration.toml /app/res/configuration.toml

CMD ["./podassrt"]
