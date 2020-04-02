Basic Kubernetes Exercises Solutions

-------------------------------------------------------------------------------------------------------------------
Exercise 0:
Step1:
(Use https://golang.org/pkg/net/http/#example_ListenAndServe as reference) 
Make new http server program, listening on port 80, which returns the current Unix time (epoch) when queried.
Solution:
// Create ListenAndServe.go file
// Contents of this file:
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

// run following commands in terminal:
go build ListenAndServe.go
go run ListenAndServe.go 

// Second command should return this
Application connecting to port 8080...

// Now open following in a browser window:
http://localhost:8080/

// It should return the Curent Unix Time
// I see this on my browser window:
Current Unix time: 1585685895

// go to terminal and press Ctrl+C to stop running the application  

Step2:
Build it with the Dockerfile from your previous exercises. 
Solution:
// Create Docker file with following contents
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

// build Dockerfile using Podman by running following in terminal
podman build -t servertag . 

// if you want to check the image, run following in terminal
podman images 

Step3:
Run the new container with podman run -p 80
Solution:
// run the container on the image by following command
podman run -p 8080:8080 --name timeServerContainer servertag

// You should see this on your terminal
Application connecting to port 8080...

// Now open following in a browser window:
http://localhost:8080/

// It should return the Curent Unix Time
// I see this on my browser window:
Current Unix time: 1585749842

// go to terminal and press Ctrl+C to stop running the application  

// to see cotainer details, keep the container running, open another terminal tab and run following
podman ps

// to see list of all running containers o the system, run the following
podman ps --all

// In case you stopped the container, and want to re-run it wih same name, remove the container and then re-run it. Use the next two commands to achieve this
podman rm -f timeServerContainer
podman run -p 8080:8080 --name timeServerContainer servertag

-------------------------------------------------------------------------------------------------------------------
Exercise 1:
Tag the new container with the server tag, and push it to your quay repository
// In terminal, login quay.io using podman
podman login quay.io

// Enter credentials (Username and Password) and you should see Login Succeeded!

// push the container to quay repo
podman push servertag  quay.io/sdhaliwa/rampup-unix-time-server-app

// Note: I ran into the following error while pushing the image 
	// Error: Error copying image to the remote destination: Error writing blob: 
	// Error initiating layer upload to /v2/sdhaliwa/rampup-unix-time-server-app/blobs/uploads/ 
	// in quay.io: unauthorized: access to the requested resource is not authorized
// I temporarily resolved it by going to quay's web console and making the rampup-unix-time-server-app repo public (find this option by going to repo -> setting icon on left vertical menu bar)
// Still trying to figure out a good solution for this error

// Try to test the image you pushed in quay repo by pulling it and running it
podman pull quay.io/sdhaliwa/rampup-unix-time-server-app
podman run -p 8080:8080 quay.io/sdhaliwa/rampup-unix-time-server-app
// Now open following in a browser window:
http://localhost:8080/

-------------------------------------------------------------------------------------------------------------------
Exercise 2:
Step1:
Spin up an OpenShift cluster.
Solution:
// Spin up a cluster by either
//1. Using installer binary(have cluster for 2 days) -- https://docs.google.com/document/d/1pWRtk7IbnfPo6cSDsopUMrxS22t3VJ2PuN39MJp9tHM/edit 
// or 2. using cluster-bot (have cluster for about 2 hours)
// go to slack and search cluster-bot app
// send this message to cluster-bot: launch
// It should send you cluster credentials in about 30-40 mins
// Once you get these cluster credentials, download the kubeconfig file
// Now go to terminal and enter following:

// export KUBECONFIG=/path/to/file

// my kubeconfig file is in my Downloads folder, so the above command for me would be:
export KUBECONFIG=/home/sdhaliwa/Downloads/cluster-bot-2020-03-23-000605.kubeconfig 

// to verify this, run following
oc cluster-info

// You can access the OpenShift web-console by the link and credentials cluster-bot provides.

Step2:
(Use https://kubernetes.io/docs/concepts/workloads/controllers/deployment as reference) 
Create a deployment which will deploy a single instance of your server
Solution:
// create deployment.yaml file
// contents of file:
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

// get a cluster
minikube start

// check kubectl is configured to talk to the cluster
kubectl version

// view nodes in cluster
kubectl get nodes

// create deployment
kubectl create deployment http-go-server --image=quay.io/sdhaliwa/rampup-unix-time-server-app
// you should see this returned: deployment.apps/http-go-server created

// get a list of deployments
kubectl get deployments

-------------------------------------------------------------------------------------------------------------------
Exercise 3:
Step1:
(Use https://kubernetes.io/docs/tutorials/stateless-application/expose-external-ip-address as reference) Expose your deployment with a Service 
Solution:
// create deployment.yaml file
// contents of file:
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

// see list of current services from cluster
kubectl get services

// create a new service and expose it to external traffic using NodePort as parameter
kubectl expose deployment/http-go-server --type="NodePort" --port 8080

Step2: Query the service from your laptop
Solution:
// see service list and get a link to target port
minikube service list
// open the target port link -- by right-clicking on it and open link
// this opens up a browser window that serves the app and shows the current unix time


// optional steps:
// clean up -- delete deployment and services
kubectl delete service http-go-server
kubectl delete deployment http-go-server
// stop minikube VM
minikube stop
// delete minikube VM
minikube delete
-------------------------------------------------------------------------------------------------------------------
