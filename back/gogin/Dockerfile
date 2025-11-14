FROM golang:1.25.4 AS gobuilder
ENV CGO_ENABLED 0
COPY . /app
WORKDIR /app
RUN go build -o gogin .

FROM scratch
COPY --from=gobuilder /app/gogin /
CMD ["/gogin"]
EXPOSE 8080