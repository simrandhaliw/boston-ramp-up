Self-help readme doc:

Steps done to solve given exercises:
_________________________________________________________________________________________
Exercise 0:
Read about and install the latest version of podman: https://podman.io/
Work through the getting started page. You can skip the checkpoint-migrate steps
Solution: (self-explanatory)

_________________________________________________________________________________________
Exercise 1:
(Use https://docs.docker.com/engine/reference/builder/ as reference)
Write a go program, which prints "Hello OpenShift" to stdout
Create a dockerfile which uses the latest golang image as a base image. The dockerfile should ensure the following happens:
- The .go file from the previous step is copied into the container
- The source is compiled into a binary
- The binary is set as the entrypoint of the container
Build a container image from the dockerfile with podman
Run the container image with podman
View the logs from the container run with podman
Solution: 
Step1] Create a go-file 
I am choosing to use vim editor to create my go file:
vi file.go

Contents of the file:
package main
import "fmt"
func main() {
 fmt.Println("Hello OpenShift!")
}

Step2] Create a dockerfile:
vi Dockerfile

Contents of the file:
FROM golang:latest AS GOAPP
WORKDIR /app 
COPY file.go file.go
RUN go build -o main .

Step3] Build a container image from dockerfile with podman
podman build --tag app -f ./Dockerfile

Step4] Run the container image with podman
podman run app

Step5] View the logs from the container run with podman
First run: podman ps -a
Then find the container id or name for the image you built.
For my image, name = brave_greider

To see logs: podman logs brave_greider
_________________________________________________________________________________________
Exercise 2:
(Use https://docs.docker.com/develop/develop-images/multistage-build/ as reference)
Do the above, but with a multi-stage build, copying the compiled binary to a second stage with ubi8 as a base image and running the executable there.
Solution:
add new content to the existing dockerfile:
vi Dockerfile

contents of file with new contents:
FROM golang:latest AS GOAPP
WORKDIR /app 
COPY file.go file.go
RUN go build -o main .

FROM ubi8  
WORKDIR /app/
COPY --from=GOAPP /app/main .
ENTRYPOINT ["./main"]

_________________________________________________________________________________________
Exercise 3:
(Use https://docs.quay.io/solution/getting-started.html as reference)
Go to quay and click sign in on the top right of the page. Click sign in with google, and log in with your Red Hat google account. This will create a quay account for you.
Go to settings and create a CLI password.
In your terminal run `podman login quay.io`
Complete the login process using your new CLI password
Push the image you created in exercise 2 to quay
Solution:
Open a web browser. Go to: https://quay.io/
Log-in with Red Hat account credentials.
Click on 'Create New Repository'.

In terminal: podman login quay.io
Enter credentials

Then run: podman tag app quay.io/sdhaliwa/rampup
(Note: your repo name path names will be different, so edit those accordingly to run above command)
Next:  podman push quay.io/sdhaliwa/rampup

Done!

To check pulling and running your app from quay.io, follow these steps:
podman pull quay.io/sdhaliwa/rampup

podman run quay.io/sdhaliwa/rampup
