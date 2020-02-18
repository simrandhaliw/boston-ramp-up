# Basic kubernetes exercises
## Exercise 0:
(Use <https://golang.org/pkg/net/http/#example_ListenAndServe> as reference)
Make new http server program, listening on port 80, which returns the current Unix time (epoch) when queried
Build it with the Dockerfile from your previous exercises
Run the new container with `podman run -p 80`

## Exercise 1:
Tag the new container with the `server` tag, and push it to your quay repository

## Exercise 2:
(Use <https://kubernetes.io/docs/concepts/workloads/controllers/deployment> as reference)
Spin up an OpenShift cluster
Create a deployment which will deploy a single instance of your server

## Exercise 3:
(Use <https://kubernetes.io/docs/tutorials/stateless-application/expose-external-ip-address> as reference)
Expose your deployment with a Service
Query the service from your laptop

