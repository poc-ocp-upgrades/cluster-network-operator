package statusmanager

import (
	"context"
	"testing"
	"time"
	configv1 "github.com/openshift/api/config/v1"
	"github.com/openshift/library-go/pkg/config/clusteroperator/v1helpers"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func init() {
	_logClusterCodePath()
	defer _logClusterCodePath()
	configv1.AddToScheme(scheme.Scheme)
	appsv1.AddToScheme(scheme.Scheme)
}
func getCO(client client.Client, name string) (*configv1.ClusterOperator, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	co := &configv1.ClusterOperator{ObjectMeta: metav1.ObjectMeta{Name: name}}
	err := client.Get(context.TODO(), types.NamespacedName{Name: name}, co)
	return co, err
}
func conditionsInclude(oldConditions, newConditions []configv1.ClusterOperatorStatusCondition) bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	for _, newCondition := range newConditions {
		foundMatchingCondition := false
		for _, oldCondition := range oldConditions {
			if newCondition.Type != oldCondition.Type || newCondition.Status != oldCondition.Status {
				continue
			}
			if newCondition.Reason != "" && newCondition.Reason != oldCondition.Reason {
				return false
			}
			if newCondition.Message != "" && newCondition.Message != oldCondition.Message {
				return false
			}
			foundMatchingCondition = true
			break
		}
		if !foundMatchingCondition {
			return false
		}
	}
	return true
}
func conditionsEqual(oldConditions, newConditions []configv1.ClusterOperatorStatusCondition) bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return conditionsInclude(oldConditions, newConditions) && conditionsInclude(newConditions, oldConditions)
}
func TestStatusManager_set(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	client := fake.NewFakeClient()
	status := New(client, "testing", "1.2.3")
	co, err := getCO(client, "testing")
	if !errors.IsNotFound(err) {
		t.Fatalf("unexpected error (expected Not Found): %v", err)
	}
	condFail := configv1.ClusterOperatorStatusCondition{Type: configv1.OperatorDegraded, Status: configv1.ConditionTrue, Reason: "Reason", Message: "Message"}
	status.Set(false, condFail)
	co, err = getCO(client, "testing")
	if err != nil {
		t.Fatalf("error getting ClusterOperator: %v", err)
	}
	if !conditionsEqual(co.Status.Conditions, []configv1.ClusterOperatorStatusCondition{condFail}) {
		t.Fatalf("unexpected Status.Conditions: %#v", co.Status.Conditions)
	}
	condProgress := configv1.ClusterOperatorStatusCondition{Type: configv1.OperatorProgressing, Status: configv1.ConditionUnknown}
	status.Set(false, condProgress)
	co, err = getCO(client, "testing")
	if err != nil {
		t.Fatalf("error getting ClusterOperator: %v", err)
	}
	if !conditionsEqual(co.Status.Conditions, []configv1.ClusterOperatorStatusCondition{condFail, condProgress}) {
		t.Fatalf("unexpected Status.Conditions: %#v", co.Status.Conditions)
	}
	condNoFail := configv1.ClusterOperatorStatusCondition{Type: configv1.OperatorDegraded, Status: configv1.ConditionFalse}
	status.Set(false, condNoFail)
	co, err = getCO(client, "testing")
	if err != nil {
		t.Fatalf("error getting ClusterOperator: %v", err)
	}
	if !conditionsEqual(co.Status.Conditions, []configv1.ClusterOperatorStatusCondition{condNoFail, condProgress}) {
		t.Fatalf("unexpected Status.Conditions: %#v", co.Status.Conditions)
	}
	condNoProgress := configv1.ClusterOperatorStatusCondition{Type: configv1.OperatorProgressing, Status: configv1.ConditionFalse}
	condAvailable := configv1.ClusterOperatorStatusCondition{Type: configv1.OperatorAvailable, Status: configv1.ConditionTrue}
	status.Set(false, condNoProgress, condAvailable)
	co, err = getCO(client, "testing")
	if err != nil {
		t.Fatalf("error getting ClusterOperator: %v", err)
	}
	if !conditionsEqual(co.Status.Conditions, []configv1.ClusterOperatorStatusCondition{condNoFail, condNoProgress, condAvailable}) {
		t.Fatalf("unexpected Status.Conditions: %#v", co.Status.Conditions)
	}
}
func TestStatusManagerSetDegraded(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	client := fake.NewFakeClient()
	status := New(client, "testing", "1.2.3")
	co, err := getCO(client, "testing")
	if !errors.IsNotFound(err) {
		t.Fatalf("unexpected error (expected Not Found): %v", err)
	}
	condFailCluster := configv1.ClusterOperatorStatusCondition{Type: configv1.OperatorDegraded, Status: configv1.ConditionTrue, Reason: "Cluster"}
	condFailOperator := configv1.ClusterOperatorStatusCondition{Type: configv1.OperatorDegraded, Status: configv1.ConditionTrue, Reason: "Operator"}
	condFailPods := configv1.ClusterOperatorStatusCondition{Type: configv1.OperatorDegraded, Status: configv1.ConditionTrue, Reason: "Pods"}
	status.SetDegraded(OperatorConfig, "Operator", "")
	co, err = getCO(client, "testing")
	if err != nil {
		t.Fatalf("error getting ClusterOperator: %v", err)
	}
	if !conditionsEqual(co.Status.Conditions, []configv1.ClusterOperatorStatusCondition{condFailOperator}) {
		t.Fatalf("unexpected Status.Conditions: %#v", co.Status.Conditions)
	}
	status.SetDegraded(ClusterConfig, "Cluster", "")
	co, err = getCO(client, "testing")
	if err != nil {
		t.Fatalf("error getting ClusterOperator: %v", err)
	}
	if !conditionsEqual(co.Status.Conditions, []configv1.ClusterOperatorStatusCondition{condFailCluster}) {
		t.Fatalf("unexpected Status.Conditions: %#v", co.Status.Conditions)
	}
	status.SetDegraded(PodDeployment, "Pods", "")
	co, err = getCO(client, "testing")
	if err != nil {
		t.Fatalf("error getting ClusterOperator: %v", err)
	}
	if !conditionsEqual(co.Status.Conditions, []configv1.ClusterOperatorStatusCondition{condFailCluster}) {
		t.Fatalf("unexpected Status.Conditions: %#v", co.Status.Conditions)
	}
	status.SetNotDegraded(OperatorConfig)
	co, err = getCO(client, "testing")
	if err != nil {
		t.Fatalf("error getting ClusterOperator: %v", err)
	}
	if !conditionsEqual(co.Status.Conditions, []configv1.ClusterOperatorStatusCondition{condFailCluster}) {
		t.Fatalf("unexpected Status.Conditions: %#v", co.Status.Conditions)
	}
	status.SetNotDegraded(ClusterConfig)
	co, err = getCO(client, "testing")
	if err != nil {
		t.Fatalf("error getting ClusterOperator: %v", err)
	}
	if !conditionsEqual(co.Status.Conditions, []configv1.ClusterOperatorStatusCondition{condFailPods}) {
		t.Fatalf("unexpected Status.Conditions: %#v", co.Status.Conditions)
	}
	status.SetNotDegraded(PodDeployment)
	co, err = getCO(client, "testing")
	if err != nil {
		t.Fatalf("error getting ClusterOperator: %v", err)
	}
	if !v1helpers.IsStatusConditionFalse(co.Status.Conditions, configv1.OperatorDegraded) {
		t.Fatalf("unexpected Status.Conditions: %#v", co.Status.Conditions)
	}
}
func TestStatusManagerSetFromDaemonSets(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	client := fake.NewFakeClient()
	status := New(client, "testing", "1.2.3")
	status.SetDaemonSets([]types.NamespacedName{{Namespace: "one", Name: "alpha"}, {Namespace: "two", Name: "beta"}})
	status.SetFromPods()
	co, err := getCO(client, "testing")
	if err != nil {
		t.Fatalf("error getting ClusterOperator: %v", err)
	}
	if !conditionsInclude(co.Status.Conditions, []configv1.ClusterOperatorStatusCondition{{Type: configv1.OperatorProgressing, Status: configv1.ConditionTrue, Reason: "Deploying"}}) {
		t.Fatalf("unexpected Status.Conditions: %#v", co.Status.Conditions)
	}
	dsA := &appsv1.DaemonSet{ObjectMeta: metav1.ObjectMeta{Namespace: "one", Name: "alpha"}}
	err = client.Create(context.TODO(), dsA)
	if err != nil {
		t.Fatalf("error creating DaemonSet: %v", err)
	}
	dsB := &appsv1.DaemonSet{ObjectMeta: metav1.ObjectMeta{Namespace: "two", Name: "beta"}}
	err = client.Create(context.TODO(), dsB)
	if err != nil {
		t.Fatalf("error creating DaemonSet: %v", err)
	}
	status.SetFromPods()
	co, err = getCO(client, "testing")
	if err != nil {
		t.Fatalf("error getting ClusterOperator: %v", err)
	}
	if !conditionsInclude(co.Status.Conditions, []configv1.ClusterOperatorStatusCondition{{Type: configv1.OperatorDegraded, Status: configv1.ConditionFalse}, {Type: configv1.OperatorProgressing, Status: configv1.ConditionTrue, Reason: "Deploying"}, {Type: configv1.OperatorAvailable, Status: configv1.ConditionFalse, Reason: "Startup"}}) {
		t.Fatalf("unexpected Status.Conditions: %#v", co.Status.Conditions)
	}
	progressingTS := metav1.Now()
	if cond := v1helpers.FindStatusCondition(co.Status.Conditions, configv1.OperatorProgressing); cond != nil {
		if cond.LastTransitionTime.IsZero() {
			t.Fatalf("progressing transition time was zero")
		}
		progressingTS = cond.LastTransitionTime
	} else {
		t.Fatalf("Progressing condition unexpectedly missing")
	}
	dsANodes := int32(1)
	dsBNodes := int32(3)
	dsA.Status.NumberUnavailable = dsANodes
	dsB.Status.NumberUnavailable = dsBNodes
	for dsA.Status.NumberUnavailable > 0 || dsB.Status.NumberUnavailable > 0 {
		err = client.Update(context.TODO(), dsA)
		if err != nil {
			t.Fatalf("error updating DaemonSet: %v", err)
		}
		err = client.Update(context.TODO(), dsB)
		if err != nil {
			t.Fatalf("error updating DaemonSet: %v", err)
		}
		status.SetFromPods()
		co, err = getCO(client, "testing")
		if err != nil {
			t.Fatalf("error getting ClusterOperator: %v", err)
		}
		if !conditionsInclude(co.Status.Conditions, []configv1.ClusterOperatorStatusCondition{{Type: configv1.OperatorProgressing, Status: configv1.ConditionTrue}, {Type: configv1.OperatorAvailable, Status: configv1.ConditionFalse}}) {
			t.Fatalf("unexpected Status.Conditions: %#v", co.Status.Conditions)
		}
		if cond := v1helpers.FindStatusCondition(co.Status.Conditions, configv1.OperatorProgressing); cond != nil {
			if !progressingTS.Equal(&cond.LastTransitionTime) {
				t.Fatalf("Progressing LastTransitionTime changed unnecessarily")
			}
		} else {
			t.Fatalf("Progressing condition unexpectedly missing")
		}
		if dsA.Status.NumberUnavailable > 0 {
			dsA.Status.NumberUnavailable--
			dsA.Status.NumberAvailable++
		}
		if dsB.Status.NumberUnavailable > 0 {
			dsB.Status.NumberUnavailable--
			dsB.Status.NumberAvailable++
		}
	}
	if dsA.Status.NumberAvailable != dsANodes || dsA.Status.NumberUnavailable != 0 || dsB.Status.NumberAvailable != dsBNodes || dsB.Status.NumberUnavailable != 0 {
		t.Fatalf("assertion failed: %#v, %#v", dsA, dsB)
	}
	err = client.Update(context.TODO(), dsA)
	if err != nil {
		t.Fatalf("error updating DaemonSet: %v", err)
	}
	err = client.Update(context.TODO(), dsB)
	if err != nil {
		t.Fatalf("error updating DaemonSet: %v", err)
	}
	time.Sleep(1 * time.Second)
	status.SetFromPods()
	co, err = getCO(client, "testing")
	if err != nil {
		t.Fatalf("error getting ClusterOperator: %v", err)
	}
	if !conditionsInclude(co.Status.Conditions, []configv1.ClusterOperatorStatusCondition{{Type: configv1.OperatorDegraded, Status: configv1.ConditionFalse}, {Type: configv1.OperatorProgressing, Status: configv1.ConditionFalse}, {Type: configv1.OperatorAvailable, Status: configv1.ConditionTrue}}) {
		t.Fatalf("unexpected Status.Conditions: %#v", co.Status.Conditions)
	}
	if cond := v1helpers.FindStatusCondition(co.Status.Conditions, configv1.OperatorProgressing); cond != nil {
		if progressingTS.Equal(&cond.LastTransitionTime) {
			t.Fatalf("Progressing LastTransitionTime didn't change when Progressing -> false")
		}
	} else {
		t.Fatalf("Progressing condition unexpectedly missing")
	}
}
func TestStatusManagerSetFromPods(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	client := fake.NewFakeClient()
	status := New(client, "testing", "1.2.3")
	status.SetDeployments([]types.NamespacedName{{Namespace: "one", Name: "alpha"}})
	status.SetFromPods()
	co, err := getCO(client, "testing")
	if err != nil {
		t.Fatalf("error getting ClusterOperator: %v", err)
	}
	if !conditionsInclude(co.Status.Conditions, []configv1.ClusterOperatorStatusCondition{{Type: configv1.OperatorProgressing, Status: configv1.ConditionTrue, Reason: "Deploying"}}) {
		t.Fatalf("unexpected Status.Conditions: %#v", co.Status.Conditions)
	}
	depB := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Namespace: "one", Name: "beta"}, Status: appsv1.DeploymentStatus{UnavailableReplicas: 1}}
	err = client.Create(context.TODO(), depB)
	if err != nil {
		t.Fatalf("error creating Deployment: %v", err)
	}
	status.SetFromPods()
	co, err = getCO(client, "testing")
	if err != nil {
		t.Fatalf("error getting ClusterOperator: %v", err)
	}
	if !conditionsInclude(co.Status.Conditions, []configv1.ClusterOperatorStatusCondition{{Type: configv1.OperatorProgressing, Status: configv1.ConditionTrue, Reason: "Deploying"}}) {
		t.Fatalf("unexpected Status.Conditions: %#v", co.Status.Conditions)
	}
	depA := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Namespace: "one", Name: "alpha"}}
	err = client.Create(context.TODO(), depA)
	if err != nil {
		t.Fatalf("error creating Deployment: %v", err)
	}
	status.SetFromPods()
	co, err = getCO(client, "testing")
	if err != nil {
		t.Fatalf("error getting ClusterOperator: %v", err)
	}
	if !conditionsInclude(co.Status.Conditions, []configv1.ClusterOperatorStatusCondition{{Type: configv1.OperatorDegraded, Status: configv1.ConditionFalse}, {Type: configv1.OperatorProgressing, Status: configv1.ConditionTrue, Reason: "Deploying"}, {Type: configv1.OperatorAvailable, Status: configv1.ConditionFalse, Reason: "Startup"}}) {
		t.Fatalf("unexpected Status.Conditions: %#v", co.Status.Conditions)
	}
	depA.Status.UnavailableReplicas = 0
	depA.Status.AvailableReplicas = 1
	err = client.Update(context.TODO(), depA)
	if err != nil {
		t.Fatalf("error updating Deployment: %v", err)
	}
	status.SetFromPods()
	co, err = getCO(client, "testing")
	if err != nil {
		t.Fatalf("error getting ClusterOperator: %v", err)
	}
	if !conditionsInclude(co.Status.Conditions, []configv1.ClusterOperatorStatusCondition{{Type: configv1.OperatorDegraded, Status: configv1.ConditionFalse}, {Type: configv1.OperatorProgressing, Status: configv1.ConditionFalse}, {Type: configv1.OperatorAvailable, Status: configv1.ConditionTrue}}) {
		t.Fatalf("unexpected Status.Conditions: %#v", co.Status.Conditions)
	}
	status.SetDeployments([]types.NamespacedName{{Namespace: "one", Name: "alpha"}, {Namespace: "one", Name: "beta"}})
	status.SetDaemonSets([]types.NamespacedName{{Namespace: "one", Name: "gamma"}})
	ds := &appsv1.DaemonSet{ObjectMeta: metav1.ObjectMeta{Namespace: "one", Name: "gamma"}, Status: appsv1.DaemonSetStatus{NumberUnavailable: 0, NumberAvailable: 1}}
	err = client.Create(context.TODO(), ds)
	if err != nil {
		t.Fatalf("error creating DaemonSet: %v", err)
	}
	status.SetFromPods()
	co, err = getCO(client, "testing")
	if err != nil {
		t.Fatalf("error getting ClusterOperator: %v", err)
	}
	if !conditionsInclude(co.Status.Conditions, []configv1.ClusterOperatorStatusCondition{{Type: configv1.OperatorDegraded, Status: configv1.ConditionFalse}, {Type: configv1.OperatorProgressing, Status: configv1.ConditionTrue}, {Type: configv1.OperatorAvailable, Status: configv1.ConditionTrue}}) {
		t.Fatalf("unexpected Status.Conditions: %#v", co.Status.Conditions)
	}
	depB.Status.UnavailableReplicas = 0
	depB.Status.AvailableReplicas = 1
	err = client.Update(context.TODO(), depB)
	if err != nil {
		t.Fatalf("error updating Deployment: %v", err)
	}
	status.SetFromPods()
	co, err = getCO(client, "testing")
	if err != nil {
		t.Fatalf("error getting ClusterOperator: %v", err)
	}
	if !conditionsInclude(co.Status.Conditions, []configv1.ClusterOperatorStatusCondition{{Type: configv1.OperatorDegraded, Status: configv1.ConditionFalse}, {Type: configv1.OperatorProgressing, Status: configv1.ConditionFalse}, {Type: configv1.OperatorAvailable, Status: configv1.ConditionTrue}}) {
		t.Fatalf("unexpected Status.Conditions: %#v", co.Status.Conditions)
	}
}
