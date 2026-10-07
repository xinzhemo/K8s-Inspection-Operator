package rules

type Condition struct {
	Field string `yaml:"field"`
	Op    string `yaml:"op"`
	Value int32  `yaml:"value"`
}
type Rule struct {
	Name      string    `yaml:"name"`
	Resource  string    `yaml:"resource"`
	Condition Condition `yaml:"condition"`
	Severity  string    `yaml:"severity"`
	Message   string    `yaml:"message"`
}
type RulesConfig struct {
	Rules []Rule `yaml:"rules"`
}
