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
							ContainerPort: 8080,
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
