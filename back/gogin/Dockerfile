FROM golang:1.25.4 AS gobuilder
ENV CGO_ENABLED 0
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

# 나머지 소스 복사
COPY . .

# 바이너리 빌드
RUN go build -o gogin .

FROM scratch
COPY --from=gobuilder /app/gogin /
CMD ["/gogin"]
EXPOSE 8080