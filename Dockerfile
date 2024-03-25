FROM golang:1.22-alpine AS build
WORKDIR /app/
COPY . .
RUN go install
RUN go build -o executable
EXPOSE 8080

FROM alpine:3.13.2
RUN apk add --no-cache tzdata
COPY --from=build /app/executable /executable
COPY --from=build /app/static /
ENV PORT=3000
EXPOSE 3000
ENTRYPOINT ["/executable"]