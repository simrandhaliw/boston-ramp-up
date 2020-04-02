FROM golang:1.9 AS GOAPP
WORKDIR /app 
COPY file.go file.go
RUN go build -o main .

FROM ubi8  
WORKDIR /app/
COPY --from=GOAPP /app/main .
ENTRYPOINT ["./main"]