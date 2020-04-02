FROM golang:1.14.1-alpine3.11 AS TimeServerApp
# Create a directory under "/" called timeServerApp
RUN mkdir /timeServerApp
# Copy files into timeServerApp Directory 
ADD . /timeServerApp
# Move to working directory /build
WORKDIR /timeServerApp 
# Build the application
RUN go build -o main .
# Export necessary port
EXPOSE 8080
# Command to run when starting the container
CMD ["/timeServerApp/main"]


