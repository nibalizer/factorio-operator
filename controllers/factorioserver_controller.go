/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	"reflect"

	"github.com/go-logr/logr"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	automatorsv1 "github.com/nibalizer/factorio-operator/api/v1"
)

// FactorioServerReconciler reconciles a FactorioServer object
type FactorioServerReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=automators.labs.nibz.science,resources=factorioservers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=automators.labs.nibz.science,resources=factorioservers/status,verbs=get;list;watch;create;update;patch;delete

func (r *FactorioServerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("factorioserver", req.NamespacedName)

	// your logic here
	var factorio automatorsv1.FactorioServer
	if err := r.Get(ctx, req.NamespacedName, &factorio); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	deployment, err := r.constructDeployment(&factorio)
	if err != nil {
		return ctrl.Result{}, err
	}

	svc, err := r.constructService(&factorio)
	if err != nil {
		return ctrl.Result{}, err
	}

	oldDeploy := &apps.Deployment{}
	objKey := client.ObjectKey{
		Namespace: factorio.Namespace,
		Name:      factorio.Name,
	}

	log.V(1).Info("getting old deployment factorio", "deployment", oldDeploy)
	if err := r.Get(ctx, objKey, oldDeploy); err != nil {
		if errors.IsNotFound(err) {
			log.V(1).Info("existing deployment not found, creating new one")
			if err := r.Create(ctx, deployment); err != nil {
				log.Error(err, "unable to create deployment for factorio", "deployment", deployment)
				return ctrl.Result{}, nil

			}
		} else {
			return ctrl.Result{}, err
		}
	} else {
		log.V(1).Info("got old deployment factorio", "deployment", oldDeploy)
		// if got deploy, do patch, else create
		trimmedDeploySpec := r.trimDeploy(&oldDeploy.Spec)
		if !reflect.DeepEqual(&deployment.Spec, trimmedDeploySpec) {
			fmt.Println(&deployment.Spec)
			fmt.Println("===")
			fmt.Println(trimmedDeploySpec)
			oldDeploy.Spec = deployment.Spec
			log.V(1).Info("Updating Deployment %s/%s\n", deployment.Namespace, deployment.Name)
			err = r.Update(context.TODO(), oldDeploy)
			if err != nil {
				return ctrl.Result{}, err
			}
		} else {
			log.V(1).Info("no change in deployment, no action")
		}
	}

	oldService := &core.Service{}
	objKey = client.ObjectKey{
		Namespace: factorio.Namespace,
		Name:      factorio.Name,
	}

	log.V(1).Info("getting old service factorio", "service", oldService)
	if err := r.Get(ctx, objKey, oldService); err != nil {
		//return ctrl.Result{}, err
		if errors.IsNotFound(err) {
			log.V(1).Info("existing service not found, creating new one")
			if err := r.Create(ctx, svc); err != nil {
				log.Error(err, "unable to create service for factorio", "service", svc)
				return ctrl.Result{}, nil

			}
		} else {
			return ctrl.Result{}, err
		}
	} else {
		log.V(1).Info("got old service factorio", "service", oldService)
		// if got deploy, do patch, else create
		trimmedServiceSpec := r.trimService(&oldService.Spec)
		if !reflect.DeepEqual(&svc.Spec, trimmedServiceSpec) {
			fmt.Println(&svc.Spec)
			fmt.Println("===")
			fmt.Println(trimmedServiceSpec)
			oldService.Spec = svc.Spec
			log.V(1).Info("Updating Service %s/%s\n", svc.Namespace, svc.Name)
			err = r.Update(context.TODO(), oldDeploy)
			if err != nil {
				return ctrl.Result{}, err
			}
		} else {
			log.V(1).Info("no change in service , no action")
		}
	}
	createdSvc := &core.Service{}
	objKey = client.ObjectKey{
		Namespace: factorio.Namespace,
		Name:      factorio.Name,
	}
	if err := r.Get(ctx, objKey, createdSvc); err != nil {
		log.Error(err, "unable to pull new service")
		//return ctrl.Result{}, err
	} else {
		if len(createdSvc.Spec.Ports) == 1 {
			fmt.Println("Service Listening on Port::", createdSvc.Spec.Ports[0].NodePort)
			factorio.Status = automatorsv1.FactorioServerStatus{}
			factorio.Status.Port = createdSvc.Spec.Ports[0].NodePort
			factorio.Status.Active = "yes"
			factorio.Status.Online = "yes"

			if err := r.Status().Update(ctx, &factorio); err != nil {
				log.Error(err, "unable to update factorio status")
				//return ctrl.Result{}, err
			}
		}
	}

	log.V(1).Info("reconciled factorio server")

	return ctrl.Result{}, nil
}

func (r *FactorioServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&automatorsv1.FactorioServer{}).
		Owns(&apps.Deployment{}).
		Owns(&core.Service{}).
		Complete(r)
}

func (r *FactorioServerReconciler) trimService(spec *core.ServiceSpec) *core.ServiceSpec {
	trimmed := core.ServiceSpec{
		Ports: []core.ServicePort{},
		Type:  spec.Type,
		Selector: map[string]string{
			"fsid": spec.Selector["fsid"],
		},
	}
	for i := range spec.Ports {
		port := core.ServicePort{
			Name:       spec.Ports[i].Name,
			Protocol:   spec.Ports[i].Protocol,
			Port:       spec.Ports[i].Port,
			TargetPort: spec.Ports[i].TargetPort,
		}
		trimmed.Ports = append(trimmed.Ports, port)
	}

	return &trimmed
}

func (r *FactorioServerReconciler) trimDeploy(spec *apps.DeploymentSpec) *apps.DeploymentSpec {

	trimmed := apps.DeploymentSpec{
		Replicas: spec.Replicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				// fsid = factorio server ID, uniuqe per instance
				"fsid": spec.Selector.MatchLabels["fsid"],
			},
		},
		Template: core.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"fsid": spec.Template.ObjectMeta.Labels["fsid"],
				},
				Annotations: make(map[string]string),
				Name:        spec.Template.ObjectMeta.Name,
				Namespace:   spec.Template.ObjectMeta.Namespace,
			},
			Spec: core.PodSpec{
				Containers: []core.Container{
					{
						Name:  spec.Template.Spec.Containers[0].Name,
						Image: spec.Template.Spec.Containers[0].Image,
					},
				},
			},
		},
	}

	return &trimmed
}

func (r *FactorioServerReconciler) constructService(factorio *automatorsv1.FactorioServer) (*core.Service, error) {
	service := &core.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"fsid":     factorio.Name,
				"gamename": "factorio",
				"version":  "1.0.3",
			},
			Annotations: make(map[string]string),
			Name:        factorio.Name,
			Namespace:   factorio.Namespace,
		},
		Spec: core.ServiceSpec{
			Ports: []core.ServicePort{
				core.ServicePort{
					Name:       "factorio",
					Protocol:   "UDP",
					Port:       34197,
					TargetPort: intstr.FromInt(34197),
				},
			},
			Selector: map[string]string{
				"fsid": factorio.Name,
			},
			Type: "NodePort",
		},
	}

	if err := ctrl.SetControllerReference(factorio, service, r.Scheme); err != nil {
		return nil, err
	}

	return service, nil

}

func (r *FactorioServerReconciler) constructDeployment(factorio *automatorsv1.FactorioServer) (*apps.Deployment, error) {

	name := factorio.Name
	podTemplate := core.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"fsid": factorio.Name,
			},
			Annotations: make(map[string]string),
			Name:        name,
			Namespace:   factorio.Namespace,
		},
		Spec: core.PodSpec{
			Containers: []core.Container{
				{
					Name:  "factorio",
					Image: "factoriotools/factorio",
				},
			},
		},
	}

	//factorio.Spec.PodSpec.Spec.Containers[0].Image = "factoriotools/factorio"
	var replicas int32
	replicas = 1

	deployment := &apps.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"fsid":     factorio.Name,
				"gamename": "factorio",
				"version":  "1.0.3",
			},
			Annotations: make(map[string]string),
			Name:        name,
			Namespace:   factorio.Namespace,
		},
		Spec: apps.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"fsid": factorio.Name,
				},
			},
			Template: podTemplate,
		},
	}

	// TODO: add cpu flags to resource (small/big/dedicated)
	// TODO: add service
	// TODO: add pvc for /opt/factorio
	// TODO: add logic for 'do we have a savefile'
	// TODO: add 'bring your own savefile' functionality

	//for k, v := range cronJob.Spec.JobTemplate.Annotations {
	//	job.Annotations[k] = v
	//}
	//job.Annotations[scheduledTimeAnnotation] = scheduledTime.Format(time.RFC3339)
	//for k, v := range cronJob.Spec.JobTemplate.Labels {
	// 	job.Labels[k] = v
	//}
	if err := ctrl.SetControllerReference(factorio, deployment, r.Scheme); err != nil {
		return nil, err
	}

	return deployment, nil
}
