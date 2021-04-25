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
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	dboperatorv1alpha1 "github.com/kabisa/db-operator/api/v1alpha1"
)

// DbReconciler reconciles a Db object
type DbReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

type DbReco struct {
	Reco
	db   dboperatorv1alpha1.Db
	dbs  map[string]DbSideDb
	conn DbServerConnectionInterface
}

func (r *DbReco) MarkedToBeDeleted() bool {
	return r.db.GetDeletionTimestamp() != nil
}

func (r *DbReco) LoadCR() (ctrl.Result, error) {
	err := r.client.Get(r.ctx, r.nsNm, &r.db)
	if err != nil {
		r.Log.Info(fmt.Sprintf("%T: %s does not exist", r.db, r.nsNm.Name))
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *DbReco) LoadObj() (bool, error) {
	var err error
	// First create conninfo without db name because we don't know whether it exists
	dbServer, err := r.GetDbServer(r.db.Spec.Server)
	if err != nil {
		r.Log.Error(err, "failed getting DbServer")
		return false, err
	}
	r.conn, err = r.GetDbConnection(dbServer, nil)
	if err != nil {
		r.Log.Error(err, "failed building dbConnection")
		return false, err
	}

	r.dbs, err = r.conn.GetDbs()
	if err != nil {
		r.Log.Error(err, "failed getting DBs")
		return false, err
	}
	_, exists := r.dbs[r.db.Spec.DbName]

	if exists {
		r.conn.Close()

		// If the database exists allow to directly adress it
		r.conn, err = r.GetDbConnection(dbServer, &r.db.Name)
		if err != nil {
			r.Log.Error(err, "failed building dbConnection")
			return false, err
		}
	}

	return exists, nil
}

func (r *DbReco) CreateObj() (ctrl.Result, error) {
	userNsName := types.NamespacedName{
		Name:      r.db.Spec.Owner,
		Namespace: r.nsNm.Namespace,
	}
	dbUser := &dboperatorv1alpha1.User{}
	err := r.client.Get(r.ctx, userNsName, dbUser)
	if err != nil {
		r.Log.Error(err, fmt.Sprintf("Failed to get User: %s", r.db.Spec.Owner))
		return ctrl.Result{}, err
	}

	r.Log.Info(fmt.Sprintf("Creating db %s", r.db.Spec.DbName))
	if r.conn == nil {
		return ctrl.Result{}, fmt.Errorf("No database connection possible")
	}
	err = r.conn.CreateDb(r.db.Spec.DbName, dbUser.Spec.UserName)
	return ctrl.Result{}, nil
}

func (r *DbReco) RemoveObj() (ctrl.Result, error) {
	if r.db.Spec.DropOnDeletion {
		r.Log.Info(fmt.Sprintf("Dropping db %s", r.db.Spec.DbName))
		err := r.conn.DropDb(r.db.Spec.DbName)
		if err != nil {
			r.Log.Error(err, fmt.Sprintf("Failed to drop db %s", r.db.Spec.DbName))
			return ctrl.Result{}, err
		}
		r.Log.Info(fmt.Sprintf("finalized db %s", r.db.Spec.DbName))
	} else {
		r.Log.Info(fmt.Sprintf("did not drop db %s as per spec", r.db.Spec.DbName))
	}
	return ctrl.Result{}, nil
}

func (r *DbReco) GetCR() client.Object {
	return &r.db
}

func (r *DbReco) EnsureCorrect() (ctrl.Result, error) {
	dbObj, _ := r.dbs[r.db.Spec.DbName]

	userNsName := types.NamespacedName{
		Name:      r.db.Spec.Owner,
		Namespace: r.nsNm.Namespace,
	}
	dbUser := &dboperatorv1alpha1.User{}
	err := r.client.Get(r.ctx, userNsName, dbUser)
	if err != nil {
		r.Log.Error(err, fmt.Sprintf("Failed to get User: %s", r.db.Spec.Owner))
		return ctrl.Result{}, err
	}
	if dbObj.Owner != dbUser.Spec.UserName {
		r.Log.Info(fmt.Sprintf("Change db %s owner to %s (%s)", dbObj.DatbaseName, r.db.Spec.Owner, dbUser.Spec.UserName))
		err = r.conn.MakeUserDbOwner(dbUser.Spec.UserName, r.db.Spec.DbName)
		if err != nil {
			r.Log.Error(err, "Failed changing db ownership")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *DbReco) CleanupConn() {
	if r.conn != nil {
		r.conn.Close()
	}
}

//+kubebuilder:rbac:groups=db-operator.kubemaster.com,resources=dbs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=db-operator.kubemaster.com,resources=dbs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=db-operator.kubemaster.com,resources=dbs/finalizers,verbs=update
func (r *DbReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("db", req.NamespacedName)
	dr := DbReco{}
	dr.Reco = Reco{r.Client, ctx, r.Log, req.NamespacedName}
	return dr.Reco.Reconcile(&dr)
}

// SetupWithManager sets up the controller with the Manager.
func (r *DbReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dboperatorv1alpha1.Db{}).
		Complete(r)
}
