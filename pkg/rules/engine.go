package rules

import (
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
)

type Anomaly struct {
	RuleName  string
	Namespace string
	PodName   string
	Severity  string
	Message   string
	Value     int32
}

func ParseRules(yamlStr string) (*RulesConfig, error) {
	var config RulesConfig
	if err := yaml.Unmarshal([]byte(yamlStr), &config); err != nil {
		return nil, err
	}
	return &config, nil
}
func MatchPod(pod corev1.Pod, rules []Rule) []Anomaly {
	var anomalies []Anomaly
	for _, rule := range rules {
		if rule.Resource != "pod" {
			continue
		}
		val, ok := extractField(pod, rule.Condition.Field)
		if !ok {
			continue
		}
		if compare(val, rule.Condition.Op, rule.Condition.Value) {
			anomalies = append(anomalies, Anomaly{
				RuleName:  rule.Name,
				Namespace: pod.Namespace,
				PodName:   pod.Name,
				Severity:  rule.Severity,
				Message:   rule.Message,
				Value:     val,
			})
		}
	}
	return anomalies
}
func extractField(pod corev1.Pod, field string) (int32, bool) {
	switch field {
	case "restartCount":
		var total int32
		for _, cs := range pod.Status.ContainerStatuses {
			total += cs.RestartCount
		}
		return total, true
	case "phase":
		switch pod.Status.Phase {
		case corev1.PodFailed:
			return 1, true
		default:
			return 0, true
		}
	}
	return 0, false
}
func compare(val int32, op string, target int32) bool {
	switch op {
	case ">":
		return val > target
	case ">=":
		return val >= target
	case "==":
		return val == target
	case "<":
		return val < target
	case "<=":
		return val <= target
	}
	return false
}
