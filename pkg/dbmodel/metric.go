package dbmodel

type MetricWhiteList struct {
	Entity
	Metric  string `json:"metric"`
	Creator string `json:"creator"`
	AppName string `json:"app_name"`
}

func (MetricWhiteList) TableName() string {
	return "metric_white_list"
}
