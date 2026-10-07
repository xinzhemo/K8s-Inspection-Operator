/*
Copyright 2026.

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

package controller

import (
	"context"
	"fmt"
	"github.com/nick/inspection-operator/pkg/metrics"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"

	inspectionv1alpha1 "github.com/nick/inspection-operator/api/v1alpha1"
	"github.com/nick/inspection-operator/pkg/cooldown"
	"github.com/nick/inspection-operator/pkg/notifier"
	"github.com/nick/inspection-operator/pkg/rules"
)

// InspectionPolicyReconciler reconciles a InspectionPolicy object
type InspectionPolicyReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	CooldownCache *cooldown.CooldownCache
}

// +kubebuilder:rbac:groups=inspection.example.com,resources=inspectionpolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=inspection.example.com,resources=inspectionpolicies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=inspection.example.com,resources=inspectionpolicies/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch
// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the InspectionPolicy object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/reconcile
func (r *InspectionPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("reconcile被触发", "namespace", req.Namespace, "name", req.Name)
	metrics.InspectionTotal.Inc()
	var policy inspectionv1alpha1.InspectionPolicy
	if err := r.Get(ctx, req.NamespacedName, &policy); err != nil {
		logger.Error(err, "无法读取inspectionpolicy")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.Info("成功读取inspectionpolicy",
		"targetNamespaces", policy.Spec.TargetNamespaces,
		"rulesConfigMap", policy.Spec.RulesConfigMap,
		"intervalSeconds", policy.Spec.IntervalSeconds,
	)
	//configmap规则
	var cm corev1.ConfigMap
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: policy.Namespace,
		Name:      policy.Spec.RulesConfigMap,
	}, &cm); err != nil {
		logger.Info("无法读取configmap")
		return ctrl.Result{}, err
	}
	//解析yaml
	rulesYaml, ok := cm.Data["rules.yaml"]
	if !ok {
		logger.Error(nil, "configmap里没有该.yaml")
		return ctrl.Result{}, nil
	}
	rulesConfig, err := rules.ParseRules(rulesYaml)
	if err != nil {
		logger.Error(err, "解析规则失败")
		return ctrl.Result{}, err
	}
	logger.Info("规则加载成功", "ruleCount", len(rulesConfig.Rules))
	for _, ns := range policy.Spec.TargetNamespaces {
		var podList corev1.PodList
		if err := r.List(ctx, &podList, client.InNamespace(ns)); err != nil {
			logger.Error(err, "无法列出Pod", "namespace", ns)
			continue
		}

		logger.Info("扫描namespace",
			"namespace", ns,
			"podCount", len(podList.Items),
		)

		for _, pod := range podList.Items {
			/*硬编码逻辑
			var totalRestarts int32
			for _, cs := range pod.Status.ContainerStatuses {
				totalRestarts += cs.RestartCount
			}

			if totalRestarts > 5 {
				logger.Info("发现pod异常重启",
					"namespace", pod.Namespace,
					"name", pod.Name,
					"restartcount", totalRestarts,
					"severity", "critical",
				)
			*/
			//利用ruleconfig.rules进行规则匹配
			anomalies := rules.MatchPod(pod, rulesConfig.Rules)
			for _, anomaly := range anomalies {
				logger.Info("发现异常",
					"namespace", anomaly.Namespace,
					"pod", anomaly.PodName,
					"rule", anomaly.RuleName,
					"severity", anomaly.Severity,
					"value", anomaly.Value,
				)
				//异常指标
				metrics.AnomalyDetectedTotal.WithLabelValues(anomaly.Namespace, anomaly.Severity).Inc()
				//冷却检查
				anomalyKey := fmt.Sprintf("%s/%s/%s", anomaly.Namespace, anomaly.PodName, anomaly.RuleName)
				cooldownDuartion := time.Duration(policy.Spec.AlertCooldown) * time.Second
				if !r.CooldownCache.ShouldAlert(anomalyKey, cooldownDuartion) {
					logger.Info("处于冷却期内，跳过告警", "key", anomalyKey)
					continue
				} else {
					if policy.Spec.NotifierWebhook != "" {
						message := fmt.Sprintf(
							"K8S巡检异常\nNamespace: %s\nPod: %s\n规则: %s\n严重等级: %s",
							anomaly.Namespace, anomaly.PodName, anomaly.Message, anomaly.Severity,
						)
						if err := notifier.SendFeishuAlert(policy.Spec.NotifierWebhook, message); err != nil {
							logger.Error(err, "飞书告警失败")
						} else {
							logger.Info("飞书告警已发送", "pod", anomaly.PodName)
						}
					}
				}
			}

			if pod.Status.Phase == corev1.PodFailed {
				logger.Info("Pod处于Failed",
					"namespace", pod.Namespace,
					"name", pod.Name,
					"phase", pod.Status.Phase,
					"severity", "critical",
				)
				metrics.AnomalyDetectedTotal.WithLabelValues(pod.Namespace, "critical").Inc()
			}
		}
	}

	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *InspectionPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&inspectionv1alpha1.InspectionPolicy{}).
		Complete(r)
}
