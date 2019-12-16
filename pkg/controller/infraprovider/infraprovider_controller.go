/* **********************************************************
 * Copyright 2019 VMware, Inc.  All rights reserved. -- VMware Confidential
 * **********************************************************/

package infraprovider

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/vmware-tanzu/vm-operator/pkg/vmprovider"
)

const (
	ControllerName = "infraprovider-controller"
)

var log = logf.Log.WithName(ControllerName)

// Add creates a new InfraProvider Controller and adds it to the Manager with default RBAC. The Manager will set fields
// on the Controller and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	// Get provider registered in the manager's main()
	provider := vmprovider.GetVmProviderOrDie()

	return &ReconcileInfraProvider{
		Client:     mgr.GetClient(),
		scheme:     mgr.GetScheme(),
		vmProvider: provider,
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New(ControllerName, mgr, controller.Options{Reconciler: r, MaxConcurrentReconciles: 1})
	if err != nil {
		return err
	}

	// As the value of CPU minimum frequency is computed based on the CPU frequencies in the nodes attached to the cluster,
	// the node addition or deletion can change this value. The event of node deletion or addition needs recomputation.
	// The create/delete node handlers enable recomputing as adding a node could lower the CPU minimum frequency, and removing
	// a node could increase the CPU minimum frequency. Updating a node or any generic events does not require recomputing CPU
	// minimum frequency because these operations do not change the node's hardware.
	// Why we have chosen node events as opposed to VC events?
	// We rely on watching node events, though we are not guaranteed that all  the hosts in the cluster are not exposed as supervisor
	// cluster nodes and those hosts could be used to run guest cluster nodes. Watching node events is okay, as that's much easier
	// to watch than VC events. If someone adds a host that doesn't join the supervisor cluster, it'd get picked up during next resync.
	infraProviderPredicate := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return true
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return false
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}

	// Watch for changes to Node
	err = c.Watch(&source.Kind{Type: &corev1.Node{}}, &handler.EnqueueRequestForObject{}, infraProviderPredicate)
	if err != nil {
		return err
	}

	return nil
}

type ReconcileInfraProvider struct {
	client.Client
	scheme     *runtime.Scheme
	vmProvider vmprovider.VirtualMachineProviderInterface
}

// Reconcile recomputes the value of cpuMinFrequency across all Hosts in the cluster. The frequency value is initialized during
// the session object creation when it is needed at the firsttime in VM controllers. The vSphere VM provider registration would
// create singleton VM provider object with SessionManger to hold per-namespace session map (per-manager/per-namespace session map
// would turn out to be per-manager session in the future to reduce the in-house management connections to vCenter ).
//
// The session objects are created as and when required and used throughout the course of controllers' reconciliation loops and available
// as long as the POD is running. This current leader would be having the latest CPU minimum frequency and all the non-leader PODs hold
// VM provider with SessionManager without having any Session objects  created.
//
//In the case of leader election usecase, the current leader would be having the latest CPU minimum frequency and all the non-leader PODs
// hold VM provider with SessionManager without having any Session objects initialized. Upon leader failover, per-namespace Session object
// gets created if reconciliation takes place. This logic still works even if we change from the per-namespace session to the single session
// in the future.
//
//The frequency is recomputed in response to the Cluster state change events. The Cluster change events that trigger the recomputation include:
//     1. Node creation
//     2. Node deletion
// The corresponding event handler enqueues the request to reconcile, and the reconcile routine carries out the frequency recomputation and updating
// operations. The updated frequency value is, in turn, used by the virtual machine controller while reconciling the virtual machines.
// +kubebuilder:rbac:groups=v1,resources=nodes,verbs=get;watch
func (r *ReconcileInfraProvider) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.Info("Received reconcile request", "namespace", request.Namespace, "name", request.Name)
	ctx := context.Background()

	err := r.vmProvider.ComputeClusterCpuMinFrequency(ctx)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}
