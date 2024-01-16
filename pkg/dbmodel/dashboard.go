package dbmodel

type Dashboard struct {
	Entity
	Name        string `json:"name"`
	Creator     string `json:"creator"`
	AppName     string `json:"app_name"`
	ChartLayout string `json:"chart_layout"`
	TagFilter   string `json:"tag_filter"`
	SavedStatus string `json:"saved_status"`
}

func (Dashboard) TableName() string {
	return "dashboard"
}

type Chart struct {
	Entity
	DashboardId uint32 `json:"dashboard_id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Aggregation string `json:"aggregation"`
	DownSample  string `json:"down_sample"`
	TopnLimit   int    `json:"topn_limit"`
}

func (Chart) TableName() string {
	return "chart"
}

type Line struct {
	Entity
	ChartId uint32 `json:"chart_id"`
	Name    string `json:"name"`
	Metric  string `json:"metric"`
	Tags    string `json:"tags"`
	Offset  int    `json:"offset"`
}

func (Line) TableName() string {
	return "line"
}

func QueryDashboardCharts(dashboardId uint32) ([]Chart, error) {
	var charts []Chart
	result := db.Where("is_deleted=? AND dashboard_id=?", 0, dashboardId).Find(&charts)
	return charts, result.Error
}

func QueryChatLines(chartId uint32) ([]Line, error) {
	var lines []Line
	result := db.Where("is_deleted=? AND chart_id=?", 0, chartId).Find(&lines)
	return lines, result.Error
}

func DeleteChartAndLines(chartId uint32) error {
	// soft delete
	result := db.Table("chart").Where("id=?", chartId).Update("is_deleted", 1)
	if result.Error != nil {
		return result.Error
	}
	result = db.Table("line").Where("chart_id=?", chartId).Update("is_deleted", 1)
	return result.Error
}

func DeleteLines(lineIds []uint32) error {
	result := db.Table("line").Where("id in ?", lineIds).Update("is_deleted", 1)
	return result.Error
}
