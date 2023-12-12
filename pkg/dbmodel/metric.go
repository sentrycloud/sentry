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

func QueryAllMetricWhiteList() ([]MetricWhiteList, error) {
	var metricWhiteList []MetricWhiteList
	result := db.Where("is_deleted=?", 0).Find(&metricWhiteList)
	return metricWhiteList, result.Error
}

func AddMetricWhiteList(metric *MetricWhiteList) error {
	metric.SetTimeNow()
	result := db.Select("metric", "creator", "app_name").Create(metric)
	return result.Error
}

func UpdateMetricWhiteList(metric *MetricWhiteList) error {
	metric.SetTimeNow()
	result := db.Model(metric).Select("metric", "creator", "app_name").Updates(metric)
	return result.Error
}

func DeleteMetricWhiteList(metric *MetricWhiteList) error {
	// soft delete
	result := db.Model(metric).Update("is_deleted", 1)
	return result.Error
}
