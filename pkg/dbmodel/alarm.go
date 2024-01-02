package dbmodel

type AlarmRule struct {
	Entity
	Name       string `json:"name"`
	Type       int    `json:"type"`
	QueryRange int    `json:"query_range"`
	Contacts   string `json:"contacts"`
	Level      int    `json:"level"`
	Message    string `json:"message"`
	DataSource string `json:"data_source"`
	Trigger    string `json:"trigger"`
}

func (AlarmRule) TableName() string {
	return "alarm_rule"
}
