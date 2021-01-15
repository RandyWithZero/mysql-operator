/*
Copyright 2021.

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
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	mysqlv1 "mysql-operator/api/v1"
)

type role string

const (
	master   = "Master"
	follower = "Follower"
	backup   = "backup"
)

// MysqlHAClusterReconciler reconciles a MysqlHACluster object
type MysqlHAClusterReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=mysql.tongtech.com,resources=mysqlhaclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mysql.tongtech.com,resources=mysqlhaclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mysql.tongtech.com,resources=mysqlhaclusters/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MysqlHACluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *MysqlHAClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("mysqlhacluster", req.NamespacedName)
	var mysqlHaCluster mysqlv1.MysqlHACluster
	if err := r.Get(ctx, req.NamespacedName, &mysqlHaCluster); err != nil {
		log.Error(err, "无法获取MysqlHACluster资源对象")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	clusterName := mysqlHaCluster.Name + "[" + mysqlHaCluster.Namespace + "]"
	log.Info("即将协调MysqlHACluster资源对象" + clusterName)
	var podList v1.PodList
	if err := r.List(ctx, &podList, client.InNamespace(req.Namespace), mysqlHaCluster.Spec.Selector); err != nil {
		log.Error(err, "无法获取MysqlHACluster联资源对象"+clusterName+"关联匹配的容器组错误")
		return ctrl.Result{}, err
	}
	var backUpQuantity uint = 0
	var followerQuantity uint = 0
    var masterPod,followerPod,backupPod *v1.Pod = nil,nil,nil
	for _, pod := range podList.Items {
		if master == pod.Annotations["role"] && (v1.PodRunning == pod.Status.Phase||v1.PodPending == pod.Status.Phase) {
			if masterPod == nil {
				masterPod=&pod
			}
		}
		if follower == pod.Annotations["role"] && (v1.PodRunning == pod.Status.Phase||v1.PodPending == pod.Status.Phase) {
			followerQuantity++
			if followerPod == nil {
				followerPod=&pod
			}
		}
		if backup == pod.Annotations["role"] && (v1.PodRunning == pod.Status.Phase||v1.PodPending == pod.Status.Phase) {
			backUpQuantity++
			if backupPod == nil {
				backupPod=&pod
			}
		}
	}
	status := &mysqlHaCluster.Status
	status.BackUpExpect = mysqlHaCluster.Spec.BackUpQuantity
	status.FollowerExpect = mysqlHaCluster.Spec.FollowerQuantity
	status.BackUpReady = backUpQuantity
	status.FollowerReady = followerQuantity

	var pod *v1.Pod = nil
	if masterPod == nil {
		status.Phase = mysqlv1.Unavailable
		if backupPod != nil {
			//todo 变更backup to master
			backupPod.Annotations["role"]=master
			log.Info("变更关联pod角色:backup to master" )
			if err := r.Update(ctx, backupPod);err!=nil{
				return ctrl.Result{}, err
			}
		} else {
			pod = makeServerPod(&mysqlHaCluster, master)
		}

	}
	if backUpQuantity != mysqlHaCluster.Spec.BackUpQuantity {
		if followerPod != nil {
			//todo  变更follower to backUp
			followerPod.Annotations["role"]=backup
			log.Info("变更关联pod角色:follower to backUp" )
			if err := r.Update(ctx, followerPod);err!=nil{
				return ctrl.Result{}, err
			}
		} else {
			pod = makeServerPod(&mysqlHaCluster, backup)
		}

	}
	if followerQuantity != mysqlHaCluster.Spec.FollowerQuantity {
		pod = makeServerPod(&mysqlHaCluster, follower)
	}
	if err := r.Status().Update(ctx, &mysqlHaCluster); err != nil {
		return ctrl.Result{}, err
	}
    //create
	if pod != nil {
		log.Info("创建关联pod:" + pod.Name)
		if err := controllerutil.SetControllerReference(&mysqlHaCluster, pod, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		if err := r.Create(ctx, pod); err != nil && !errors.IsAlreadyExists(err) {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

func makeServerPod(mysql *mysqlv1.MysqlHACluster, role string) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        mysql.Name + "-" + mysql.Spec.MysqlServer.Name + "",
			Namespace:   mysql.Namespace,
			Annotations: map[string]string{"role": role},
		},
		Spec: *mysql.Spec.MysqlServer.Spec.DeepCopy(),
	}
}
// SetupWithManager sets up the controller with the Manager.
func (r *MysqlHAClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	c := ctrl.NewControllerManagedBy(mgr)
	c.Watches(&source.Kind{Type: &v1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &mysqlv1.MysqlHACluster{},
	})
	return c.For(&mysqlv1.MysqlHACluster{}).
		Complete(r)
}
