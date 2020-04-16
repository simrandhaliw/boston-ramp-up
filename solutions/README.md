
------------ 

## Exercise 0:
Read about and install the latest version of podman: https://podman.io/
Work through the getting started page. You can skip the checkpoint-migrate steps

### Solution 0: 
(self-explanatory)

------------ 

## Exercise 1:
(Use https://docs.docker.com/engine/reference/builder/ as reference)
Write a go program, which prints "Hello OpenShift" to stdout
Create a dockerfile which uses the latest golang image as a base image. The dockerfile should ensure the following happens:
- The .go file from the previous step is copied into the container
- The source is compiled into a binary
- The binary is set as the entrypoint of the container
Build a container image from the dockerfile with podman
Run the container image with podman
View the logs from the container run with podman

### Solution 1: 
1. Create a .go file 

```
// file.go
package main
import "fmt"
func main() {
 fmt.Println("Hello OpenShift!")
}
```

2. Create a dockerfile

```
# Dockerfile
FROM golang:latest AS GOAPP
WORKDIR /app 
COPY file.go file.go
RUN go build -o main .
```

3. Build a container image from dockerfile with podman

```
podman build --tag app -f ./Dockerfile
```

Note: refer podman basic commands from [Podman Cheatsheet](https://developers.redhat.com/cheat-sheets/podman-basics/)

4. Run the container image with podman

```
podman run app
```

5. View the logs from the container run with podman

```
podman ps 
```

Note: If you see nothing returned from this command, it is possible that your container has stopped working. 
Try running the app again, and then open another terminal tab/window and try running the above command.

or to see the logs, you can also run 

```
podman ps --all
```

Then find the container id or name for the image you built.
For exmple my image name = brave_greider

To see logs run

```
podman logs brave_greider
```

------------ 

## Exercise 2:
(Use https://docs.docker.com/develop/develop-images/multistage-build/ as reference)
Do the above, but with a multi-stage build, copying the compiled binary to a second stage with ubi8 as a base image and running the executable there.

###  Solution 2:
Replace the existing Dockerfile content with the following content

```
FROM golang:latest AS GOAPP
WORKDIR /app 
COPY file.go file.go
RUN go build -o main .

FROM ubi8  
WORKDIR /app/
COPY --from=GOAPP /app/main .
ENTRYPOINT ["./main"]
```

------------ 

## Exercise 3:
(Use https://docs.quay.io/solution/getting-started.html as reference)
Go to quay and click sign in on the top right of the page. Click sign in with google, and log in with your Red Hat google account. This will create a quay account for you.
Go to settings and create a CLI password.
In your terminal run `podman login quay.io`
Complete the login process using your new CLI password
Push the image you created in exercise 2 to quay

###  Solution 3:

```
podman login quay.io
```
Enter your quay.io account credentials

```
podman tag app quay.io/sdhaliwa/rampup
```
Note: I chose my repository name as 'rampup'. You can choose any name.

```
podman push quay.io/sdhaliwa/rampup
```

Done!

To check if your app was successfully pushed to the quay.io repo, try pulling and running your app:

```
podman pull quay.io/sdhaliwa/rampup
```

```
podman run quay.io/sdhaliwa/rampup
```
------------ 
