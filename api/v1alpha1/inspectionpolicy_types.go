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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// InspectionPolicySpec 定义巡检策略的期望状态
type InspectionPolicySpec struct {
	// 巡检哪些 namespace
	// +kubebuilder:validation:MinItems=1
	TargetNamespaces []string `json:"targetNamespaces"`

	// 巡检规则所在的 ConfigMap 名称
	RulesConfigMap string `json:"rulesConfigMap"`

	// 巡检周期（秒）
	// +kubebuilder:validation:Minimum=10
	// +kubebuilder:default=60
	IntervalSeconds int `json:"intervalSeconds,omitempty"`

	// 同一异常告警冷却时间（秒）
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=300
	AlertCooldown int `json:"alertCooldown,omitempty"`

	// 飞书 Webhook 地址
	NotifierWebhook string `json:"notifierWebhook,omitempty"`
}

// InspectionPolicyStatus 定义巡检的实际状态
type InspectionPolicyStatus struct {
	// 最近一次巡检时间
	LastInspectionTime *metav1.Time `json:"lastInspectionTime,omitempty"`

	// 累计巡检次数
	TotalInspections int64 `json:"totalInspections,omitempty"`

	// 累计发现异常数
	TotalAnomalies int64 `json:"totalAnomalies,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// InspectionPolicy 定义巡检策略资源
type InspectionPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InspectionPolicySpec   `json:"spec,omitempty"`
	Status InspectionPolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// InspectionPolicyList 是 InspectionPolicy 的列表
type InspectionPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InspectionPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&InspectionPolicy{}, &InspectionPolicyList{})
}
