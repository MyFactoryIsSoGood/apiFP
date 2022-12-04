FROM golang:latest

ENV GOPATH=/

COPY ./ ./

ENV BASE_URL="http:/0.0.0.0:8080"
ENV PORT=":8080"
ENV DB_USERNAME="postgres"
ENV DB_NAME="Fingerprints"
ENV DB_PASSWORD="123qwe123"

RUN go mod download
RUN go build -o fp ./main.go

EXPOSE 8080
CMD ["./fp"]