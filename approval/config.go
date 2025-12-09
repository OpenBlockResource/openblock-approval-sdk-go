package approval

type Config struct {
	Role           string                 `yaml:"role"`
	APIKey         string                 `yaml:"apiKey"`
	APISecret      string                 `yaml:"apiSecret"`
	ApprovalParams []ApprovalParam        `yaml:"approvalParams"`
	TxInfo         map[string]interface{} `yaml:"txInfo"`
}

type ApprovalParam struct {
	MatchParams  []Param `yaml:"matchParams"`
	VerifyParams []Param `yaml:"verifyParams"`
}

type Param struct {
	Path  string `yaml:"path"`
	Value string `yaml:"value"`
	Rule  string `yaml:"rule"`
}
