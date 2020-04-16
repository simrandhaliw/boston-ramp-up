
---------------
## Exercise 0:
(Use https://golang.org/pkg/net/http/#example_ListenAndServe as reference) 
Make new http server program, listening on port 80, which returns the current Unix time (epoch) when queried. Build it with the Dockerfile from your previous exercises. Run the new container with podman run -p 80.

### Solution 0:

1. 

Create a dir to work in

```
mkdir k8s-solutions
```
```
cd k8s-solutions
```

Create ListenAndServe.go file

```
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

// main function: new http server program, listening on port 80,
// to return current Unix time (epoch) when queried
func main() {
	timeHandler := func(w http.ResponseWriter, req *http.Request) {
		currentTime := time.Now().Unix()
		io.WriteString(w, "Current Unix time: "+strconv.Itoa(int(currentTime)))
	}
	fmt.Println("Application connecting to port 8080...")
	http.HandleFunc("/", timeHandler) 
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

Run following commands in terminal

```
go build ListenAndServe.go
go run ListenAndServe.go 
```

Second command should return this

```
Application connecting to port 8080...
```

Now open following in a browser window:

http://localhost:8080/

It should return the Curent Unix Time.
Example: 

```
Current Unix time: 1585685895
```

To stop application return to terminal and press ```Ctrl+C``` to stop running the application.  

2. 

Create a Dockerfile

```
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
```	

Build Dockerfile using Podman by running following in terminal

```
podman build -t servertag . 
```
Note: servertag can be replaced with any tag name of your choice

To check the image, run following in terminal

```
podman images 
```

3. 

Run the container on the image by following command

```
podman run -p 8080:8080 --name timeServerContainer servertag
```
Note: timeServerContainer can be replaced with any container name of your choice

You should see this on your terminal
```
Application connecting to port 8080...
```

Now open following in a browser window:
http://localhost:8080/

It should return the Curent Unix Time. Example:
```
Current Unix time: 1585749842
```

To see container details, keep the container running, open another terminal tab and run following
```
podman ps
```

To see list of all running containers on the system, run the following

```
podman ps --all
```

In case you stopped the container, and want to re-run it wih same name, remove the container and then re-run it. Use the next two commands to achieve this
```
podman rm -f timeServerContainer
podman run -p 8080:8080 --name timeServerContainer servertag
```
---------------

## Exercise 1:
Tag the new container with the server tag, and push it to your quay repository

### Solution 1:
In terminal, login quay.io using podman
```
podman login quay.io
```
Enter credentials (Username and Password) and you should see `Login Succeeded!`

Push the container to quay repo
```
podman push servertag  quay.io/sdhaliwa/rampup-unix-time-server-app
```

Note: I ran into the following error while pushing the image 

`
Error: Error copying image to the remote destination: Error writing blob: 
Error initiating layer upload to /v2/sdhaliwa/rampup-unix-time-server-app/blobs/uploads/ 
in quay.io: unauthorized: access to the requested resource is not authorized
`	

 I temporarily resolved it by going to quay's web console and making the rampup-unix-time-server-app repo public (find this option by going to repo -> setting icon on left vertical menu bar)


Try to test the image you pushed in quay repo by pulling it and running it

```
podman pull quay.io/sdhaliwa/rampup-unix-time-server-app
podman run -p 8080:8080 quay.io/sdhaliwa/rampup-unix-time-server-app
```

Now open following in a browser window:
http://localhost:8080/

And it should return the Curent Unix Time. Example:
```
Current Unix time: 1585749842
```


---------------

## Exercise 2:
Spin up an OpenShift cluster. (Use https://kubernetes.io/docs/concepts/workloads/controllers/deployment as reference) 
Create a deployment which will deploy a single instance of your server.
### Solution 2:

1.

Spin up a cluster by either

1st method: 
Using installer binary(have cluster for 2 days). 
Visit this doc for more details:
https://docs.google.com/document/d/1pWRtk7IbnfPo6cSDsopUMrxS22t3VJ2PuN39MJp9tHM/edit 

2nd method: 
Using cluster-bot (have cluster for about 2 hours)

Go to slack and search for `cluster-bot` app.
Send this message to cluster-bot: `launch`.

It should send you cluster credentials in about 30-40 mins.
Once you get these cluster credentials, download the kubeconfig file.
Now go to terminal and enter following:

```
export KUBECONFIG=/path/to/file
```

My kubeconfig file is in my Downloads folder, so the above command for me would be:

```
export KUBECONFIG=/home/sdhaliwa/Downloads/cluster-bot-2020-03-23-000605.kubeconfig 
```
To verify this, run following
```
oc cluster-info
```

You can access the OpenShift web-console by the link and credentials cluster-bot provides.

2.

Create deployment.yaml file

```
apiVersion: apps/v1
kind: Deployment
metadata:
	name: http-go-server
	labels:
	app: serverapp
spec:
	replicas: 3
	selector:
	matchLabels:
		app: serverapp
	template:
	metadata:
		labels:
	app: serverapp
	spec:
		containers:
		- name: serverapp
	image: quay.io/sdhaliwa/rampup-unix-time-server-app
	ports:
	- containerPort: 8080    
```		    

To test out your code, I would recommend create a local cluster using minikube (instead of an AWS cluster) 
```
minikube start
```

Check if  `kubectl` is configured to talk to the cluster
```
kubectl version
```

View nodes in cluster
```
kubectl get nodes
```

Create deployment
```
kubectl create deployment http-go-server --image=quay.io/sdhaliwa/rampup-unix-time-server-app
```
Get a list of deployments
```
kubectl get deployments
```

---------------

## Exercise 3:

(Use https://kubernetes.io/docs/tutorials/stateless-application/expose-external-ip-address as reference) Expose your deployment with a Service.  Query the service from your laptop.

### Solution 3:
Create deployment.yaml file

```
apiVersion: v1
kind: Service
metadata:
	name: app-service
	labels:
	app: serverapp
spec: 
	selector:
	app: http-go-server
	ports:
	- protocol: TCP
	port: 8080
	targetPort: 8080
	name: goserver
```		

See list of current services from cluster
```
kubectl get services
```

Create a new service and expose it to external traffic using NodePort as parameter
```
kubectl expose deployment/http-go-server --type="NodePort" --port 8080
```

2.

See service list and get a link to target port
```
minikube service list
```

Open the target port link by right-clicking on it.
This opens up a browser window that serves the app and shows the current unix time.

Done!

Clean up by delete deployment and services (optional steps)
```
kubectl delete service http-go-server
```
```
kubectl delete deployment http-go-server
```

Stop minikube VM
```
minikube stop
```

Delete minikube VM
```
minikube delete
```

---------------
