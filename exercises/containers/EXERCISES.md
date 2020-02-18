# Basic Go and container exercises
### For each of the reference docs, use podman instead of docker

## Exercise 0:
Read about and install the latest version of podman: <https://podman.io/>
Work through the getting started page. You can skip the checkpoint-migrate steps

## Exercise 1:
(Use <https://docs.docker.com/engine/reference/builder> as reference)
Write a go program, which prints "Hello OpenShift" to stdout
Create a dockerfile which uses the latest golang image as a base image. The dockerfile should ensure the following happens:
- The .go file from the previous step is copied into the container
- The source is compiled into a binary
- The binary is set as the entrypoint of the container
Build a container image from the dockerfile with podman
Run the container image with podman
View the logs from the container run with podman

## Exercise 2:
(Use <https://docs.docker.com/develop/develop-images/multistage-build> as reference)
Do the above, but with a multi-stage build, copying the compiled binary to a second stage with ubi8 as a base image and running the executable there.

## Exercise 3:
(Use <https://docs.quay.io/solution/getting-started.html> as reference)
Go to quay and click sign in on the top right of the page. Click sign in with google, and log in with your Red Hat google account. This will create a quay account for you.
Go to settings and create a CLI password.
In your terminal run `podman login quay.io`
Complete the login process using your new CLI password
Push the image you created in exercise 2 to quay

