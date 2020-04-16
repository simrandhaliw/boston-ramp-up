
-------------
## Assignment
Create a go based operator using the Operator SDK https://github.com/operator-framework/operator-sdk
This operator should be based around the TimeServer CRD which you will create.
The controller for TimeServer should create a HTTP server that returns the current time when queried.
One server should be created for each replica specified in the CRD. Each TimeServer should exist on a different
worker node. If there are more replicas specified than worker nodes, only create as many TimeServers as nodes.
When the amount of replicas in the CRD is changed, the controller should change the number of TimeServers if possible.

### Solution:
A good starting point to start this is by reading this:
https://github.com/operator-framework/getting-started

Install operator SDK CLI by following:
https://github.com/operator-framework/operator-sdk/blob/master/doc/user/install-operator-sdk.md

Follow steps from the following guides:
https://github.com/operator-framework/operator-sdk and
https://github.com/operator-framework/operator-sdk/blob/master/doc/user-guide.md

1.
Just like we built an app to display current Unix time when queried in Exercise 2, you need to built an app that displays current time a specific timezone.

Follow the steps in Exercise 2, just replace the .go file with this:
<details>
	<summary>.go file contents</summary>
		
		// name of file: TimeZoneListenAndServe.go
		package main

		import (
			"fmt"
			"io/ioutil"
			"log"
			"net/http"
			"strings"
			"time"
		)

		// Get a list of valid timezones. See this link for help:
		// https://stackoverflow.com/questions/40120056/get-a-list-of-valid-time-zones-in-go
		var zoneDirs = []string{
			// Update path according to your OS
			"/usr/share/zoneinfo/",
			"/usr/share/lib/zoneinfo/",
			"/usr/lib/locale/TZ/",
		}

		var zoneDir string
		var listLocale []string

		func readFile(path string) {
			files, _ := ioutil.ReadDir(zoneDir + path)
			for _, f := range files {
				if f.Name() != strings.ToUpper(f.Name()[:1])+f.Name()[1:] {
					continue
				}
				if f.IsDir() {
					readFile(path + "/" + f.Name())
				} else {
					localeVal := (path + "/" + f.Name())[1:]
					_, err := time.LoadLocation(localeVal)
					if err != nil {
						log.Fatal(err)
					}
					listLocale = append(listLocale, localeVal)
					// fmt.Println((path + "/" + f.Name())[1:])
				}
			}
		}

		func servingHandler(locale string) func(http.ResponseWriter, *http.Request) {
			return func(w http.ResponseWriter, r *http.Request) {
				loc, err := time.LoadLocation(locale)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Fprint(w, "Time time is "+time.Now().In(loc).Format("2006-01-02 15:04:05 PM"))
			}
		}

		func timeHandler(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Select one of the folowing timezones and add that to end of query string\n")
			for _, val := range listLocale {
				fmt.Fprintf(w, "%s\n", val)
			}
		}

		// main function: new http server program, listening on port 80,
		// to return time (epoch) when queried
		func main() {
			for _, zoneDir = range zoneDirs {
				readFile("")
			}

			http.HandleFunc("/", timeHandler)
			for _, val := range listLocale {
				http.HandleFunc("/"+val, servingHandler(val))
			}
			fmt.Println("Application connecting to port 8080...")
			log.Fatal(http.ListenAndServe(":8080", nil))
		}

</details>

and replace Dockerfile with this:

<details>
	<summary>Dokerfile contents</summary>

		FROM golang:1.12 AS GOAPP
		WORKDIR /src/
		COPY TimeZoneListenAndServe.go main.go
		RUN go build -o app .
		FROM fedora
		WORKDIR /exec/
		COPY --from=GOAPP /src/app .
		ENTRYPOINT ["./app"]
		EXPOSE 8080	

</details>	

Follow rest of the steps in a similar manner.

For me this app was pushed to this quay.io repo: `quay.io/sdhaliwa/time-zone-server-app`

2. 
Create a new github repo. 

Mine is: github.com/simrandhaliw/time-server-app-operator

3. 

Create an operator using Operator-sdk

```
cd /home/sdhaliwa/dev/git/openshift/

mkdir operator-project

cd operator-project

operator-sdk new time-server-operator --type go --repo github.com/simrandhaliw/time-server-app-operator

cd time-server-operator
```
Do `ls` and you would see these folders/files: `build  cmd  deploy  go.mod  go.sum  pkg  tools.go  version`

Now since you are running this outside of GOPATH, run this:
```
export GO111MODULE=on
``` 

```
operator-sdk add api --kind TimeServer --api-version timeserver.example.com/v1alpha1
```
Running above command dispayed this error:
`INFO[0000] Running deepcopy code-generation for Custom Resource group versions: [timeserver:[v1alpha1], ] 
F0404 22:51:14.590293  776609 deepcopy.go:885] Hit an unsupported type invalid type for invalid type, from ./pkg/apis/timeserver/v1alpha1.TimeServer`

I checked if GOPATH and GOROOT were set properly by running 

```
go env
```

but again, error persisted. So I manually set GOROOT again

```
export GOROOT=/usr/local/go
```

This time the following command ran successfully
```
'operator-sdk add api --kind TimeServer --api-version timeserver.example.com/v1alpha1' 
```

Using any code editor (I used vscode):

Go to `pkg` --> `apis` --> `timeserver` --> `v1alpha1` --> `timeserver_types.go`, and add the following:

```
type TimeServerSpec struct {
	// Size is the size of the memcached deployment
	Size int32 `json:"size"`
}
type TimeServerStatus struct {
	// Nodes are the names of the memcached pods
	Nodes []string `json:"nodes"`
}
```	

After modifying `*_types.go`, always run following command to update the generated code for that resource type
```
operator-sdk generate k8s
```

Update CRD manifests
```
operator-sdk generate crds
```

Add new controller 
```
operator-sdk add controller  --kind TimeServer --api-version timeserver.example.com/v1alpha1
```

Now replace `timeserver_controller.go` (`pkg` --> `controller` --> `timeserver` --> `timeserver_controller.go`) with following content

<details>
 <summary>Click to expand the code content</summary>
		
		package timeserver

		import (
			"context"
			"fmt"
			"reflect"

			timeserverv1alpha1 "github.com/simrandhaliw/time-server-app-operator/pkg/apis/timeserver/v1alpha1"

			appsv1 "k8s.io/api/apps/v1"
			corev1 "k8s.io/api/core/v1"
			"k8s.io/apimachinery/pkg/api/errors"
			metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
			"k8s.io/apimachinery/pkg/runtime"
			"k8s.io/apimachinery/pkg/types"
			"k8s.io/apimachinery/pkg/util/intstr"
			"sigs.k8s.io/controller-runtime/pkg/client"
			"sigs.k8s.io/controller-runtime/pkg/controller"
			"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
			"sigs.k8s.io/controller-runtime/pkg/handler"
			logf "sigs.k8s.io/controller-runtime/pkg/log"
			"sigs.k8s.io/controller-runtime/pkg/manager"
			"sigs.k8s.io/controller-runtime/pkg/reconcile"
			"sigs.k8s.io/controller-runtime/pkg/source"
		)

		var log = logf.Log.WithName("controller_timeserver")

		/**
		* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
		* business logic.  Delete these comments after modifying this file.*
			*/

		// Add creates a new TimeServer Controller and adds it to the Manager. The Manager will set fields on the Controller
		// and Start it when the Manager is Started.
		func Add(mgr manager.Manager) error {
			return add(mgr, newReconciler(mgr))
		}

		// newReconciler returns a new reconcile.Reconciler
		func newReconciler(mgr manager.Manager) reconcile.Reconciler {
			return &ReconcileTimeServer{client: mgr.GetClient(), scheme: mgr.GetScheme()}
		}

		// add adds a new Controller to mgr with r as the reconcile.Reconciler
		func add(mgr manager.Manager, r reconcile.Reconciler) error {
			// Create a new controller
			c, err := controller.New("timeserver-controller", mgr, controller.Options{Reconciler: r})
			if err != nil {
				return err
			}

			// Watch for changes to primary resource TimeServer
			err = c.Watch(&source.Kind{Type: &timeserverv1alpha1.TimeServer{}}, &handler.EnqueueRequestForObject{})
			if err != nil {
				return err
			}

			// TODO(user): Modify this to be the types you create that are owned by the primary resource
			// Watch for changes to secondary resource Pods and requeue the owner TimeServer
			err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
				IsController: true,
				OwnerType:    &timeserverv1alpha1.TimeServer{},
			})
			if err != nil {
				return err
			}

			return nil
		}

		// blank assignment to verify that ReconcileTimeServer implements reconcile.Reconciler
		var _ reconcile.Reconciler = &ReconcileTimeServer{}

		// ReconcileTimeServer reconciles a TimeServer object
		type ReconcileTimeServer struct {
			// TODO: Clarify the split client
			// This client, initialized using mgr.Client() above, is a split client
			// that reads objects from the timeserver and writes to the apiserver
			client client.Client
			scheme *runtime.Scheme
		}

		// Reconcile reads that state of the cluster for a TimeServer object and makes changes based on the state read
		// and what is in the TimeServer.Spec
		// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
		// a TimeServer Deployment for each TimeServer CR
		// Note:
		// The Controller will requeue the Request to be processed again if the returned error is non-nil or
		// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
		func (r *ReconcileTimeServer) Reconcile(request reconcile.Request) (reconcile.Result, error) {
			reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
			reqLogger.Info("Reconciling TimeServer")

			// Fetch the TimeServer instance
			timeserver := &timeserverv1alpha1.TimeServer{}
			err := r.client.Get(context.TODO(), request.NamespacedName, timeserver)
			if err != nil {
				if errors.IsNotFound(err) {
					// Request object not found, could have been deleted after reconcile request.
					// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
					// Return and don't requeue
					reqLogger.Info("TimeServer resource not found. Ignoring since object must be deleted")
					return reconcile.Result{}, nil
				}
				// Error reading the object - requeue the request.
				reqLogger.Error(err, "Failed to get TimeServer")
				return reconcile.Result{}, err
			}

			// Check if the deployment already exists, if not create a new one
			deploymentFound := &appsv1.Deployment{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: timeserver.Name, Namespace: timeserver.Namespace}, deploymentFound)
			if err != nil && errors.IsNotFound(err) {
				// Define a new deployment
				dep := r.deploymentForTimeServer(timeserver)
				reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
				err = r.client.Create(context.TODO(), dep)
				if err != nil {
					reqLogger.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
					return reconcile.Result{}, err
				}
				// Deployment created successfully - return and requeue
				return reconcile.Result{Requeue: true}, nil
			} else if err != nil {
				reqLogger.Error(err, "Failed to get Deployment")
				return reconcile.Result{}, err
			}

			// Check if the service already exists, if not create a new one
			serviceFound := &corev1.Service{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: timeserver.Name, Namespace: timeserver.Namespace}, serviceFound)
			if err != nil && errors.IsNotFound(err) {
				// Define a new service
				ser := r.serviceForTimeServer(timeserver)
				reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", ser.Namespace, "Deployment.Name", ser.Name)
				err = r.client.Create(context.TODO(), ser)
				if err != nil {
					reqLogger.Error(err, "Failed to create new Deployment", "Deployment.Namespace", ser.Namespace, "Deployment.Name", ser.Name)
					return reconcile.Result{}, err
				}
				// Deployment created successfully - return and requeue
				return reconcile.Result{Requeue: true}, nil
			} else if err != nil {
				reqLogger.Error(err, "Failed to get Deployment")
				return reconcile.Result{}, err
			}

			// Ensure the deployment size is the same as the spec
			size := timeserver.Spec.Size
			if *deploymentFound.Spec.Replicas != size {
				deploymentFound.Spec.Replicas = &size
				err = r.client.Update(context.TODO(), deploymentFound)
				if err != nil {
					reqLogger.Error(err, "Failed to update Deployment", "Deployment.Namespace", deploymentFound.Namespace, "Deployment.Name", deploymentFound.Name)
					return reconcile.Result{}, err
				}
				// Spec updated - return and requeue
				return reconcile.Result{Requeue: true}, nil
			}

			// Update the TimeServer status with the pod names
			// List the pods for this timeserver's deployment
			podList := &corev1.PodList{}
			listOpts := []client.ListOption{
				client.InNamespace(timeserver.Namespace),
				client.MatchingLabels(labelsForTimeServer(timeserver.Name)),
			}
			if err = r.client.List(context.TODO(), podList, listOpts...); err != nil {
				reqLogger.Error(err, "Failed to list pods", "TimeServer.Namespace", timeserver.Namespace, "TimeServer.Name", timeserver.Name)
				return reconcile.Result{}, err
			}
			podNames := getPodNames(podList.Items)

			// Update status.Nodes if needed
			if !reflect.DeepEqual(podNames, timeserver.Status.Nodes) {
				timeserver.Status.Nodes = podNames
				err := r.client.Status().Update(context.TODO(), timeserver)
				if err != nil {
					reqLogger.Error(err, "Failed to update TimeServer status")
					return reconcile.Result{}, err
				}
			}

			return reconcile.Result{}, nil
		}

		// deploymentForTimeServer returns a timeserver Deployment object
		func (r *ReconcileTimeServer) deploymentForTimeServer(m *timeserverv1alpha1.TimeServer) *appsv1.Deployment {
			ls := labelsForTimeServer(m.Name)
			replicas := m.Spec.Size

			dep := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      m.Name,
					Namespace: m.Namespace,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: &replicas,
					Selector: &metav1.LabelSelector{
						MatchLabels: ls,
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: ls,
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{
								Image: "quay.io/sdhaliwa/time-zone-server-app",
								Name:  "timeserver",
								Ports: []corev1.ContainerPort{{
									ContainerPort: 11211,
									Name:          "timeserver",
								}},
							}},
						},
					},
				},
			}
			// Set TimeServer instance as the owner and controller of Deployment
			err := controllerutil.SetControllerReference(m, dep, r.scheme)
			if err != nil {
				log.Error(err, fmt.Sprintf("Error when trying to set TimeServer instance as the owner and controller of Deployment"))
			}
			return dep
		}

		// serviceForTimeServer returns a timeserver Sevice object
		func (r *ReconcileTimeServer) serviceForTimeServer(m *timeserverv1alpha1.TimeServer) *corev1.Service {
			ls := labelsForTimeServer(m.Name)

			ser := &corev1.Service{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Service",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "service-" + m.Name,
					Namespace: m.Namespace,
					Labels:    ls,
				},
				Spec: corev1.ServiceSpec{
					Type:     corev1.ServiceTypeClusterIP,
					Selector: ls,
					Ports: []corev1.ServicePort{
						{
							Name:       "http",
							Port:       8080,
							TargetPort: intstr.FromInt(8080),
						},
					},
				},
			}
			// Set TimeServer instance as the owner and controller of Service
			err := controllerutil.SetControllerReference(m, ser, r.scheme)
			if err != nil {
				log.Error(err, fmt.Sprintf("Error when trying to set TimeServer instance as the owner and controller of Service"))
			}
			return ser
		}

		// labelsForTimeServer returns the labels for selecting the resources
		// belonging to the given timeserver CR name.
		func labelsForTimeServer(name string) map[string]string {
			return map[string]string{"app": "timeserver", "timeserver_cr": name}
		}

		// getPodNames returns the pod names of the array of pods passed in
		func getPodNames(pods []corev1.Pod) []string {
			var podNames []string
			for _, pod := range pods {
				podNames = append(podNames, pod.Name)
			}
			return podNames
		}
		
</details>


Register CRD with k8s apiserver
```
kubectl create -f deploy/crds/timeserver.example.com_timeservers_crd.yaml
```

Login to quay.io using docker
```
docker login quay.io
```

Build the operator image 
```
operator-sdk build quay.io/sdhaliwa/time-server-app-operator
```

Push this image to quay
```
docker push quay.io/sdhaliwa/time-server-app-operator
```

Change operator image in operator.yaml to this
```
image: quay.io/sdhaliwa/time-server-app-operator
```
4.

Setup Service Account
```
kubectl create -f deploy/service_account.yaml
```

Setup RBAC
```
kubectl create -f deploy/role.yaml
```
```
kubectl create -f deploy/role_binding.yaml
```

Setup the CRD
```
kubectl create -f deploy/crds/timeserver.example.com_timeservers_crd.yaml
```

Deploy the app-operator
```
kubectl create -f deploy/operator.yaml
```

Create an AppService CR

The default controller will watch for AppService objects and create a pod for each CR
```
kubectl create -f deploy/crds/timeserver.example.com_v1alpha1_timeserver_cr.yaml
```

Create a service
```
kubectl expose deployment/example-timeserver --type="NodePort" --port 8080
```

```
kubectl get pods
kubectl get nodes
kubectl get deployments
kubectl get services
```
```
minikube service list
```
or 
```
minikube service example-timeserver --url
```

Note: While testing things out, I had to delete the following resources and then make changes and create them again. Just in case you want to do the same, the following would prove useful to just copy-paste.
```
kubectl delete -f deploy/crds/timeserver.example.com_v1alpha1_timeserver_cr.yaml
kubectl delete -f deploy/operator.yaml
kubectl delete -f deploy/role.yaml
kubectl delete -f deploy/role_binding.yaml
kubectl delete -f deploy/service_account.yaml
kubectl delete -f deploy/crds/timeserver.example.com_timeservers_crd.yaml
kubectl delete service <service-name>
```

-------------
## Bonus
Watch the node objects in the cluster as well. When the number of nodes change, adjust the amount deployed TimeServers
if necessary

-------------
## Super Stretch Bonus
Write end to end tests to test the functionality of your operator. Your tests should alter the TimeServer replicas
and ensure that the actual TimeServers reflect that change. You should do the same with the node count. You should
also ensure that each node has exactly 1 or 0 TimeServers deployed on it.

-------------
