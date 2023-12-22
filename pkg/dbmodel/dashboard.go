package dbmodel

type Dashboard struct {
	Entity
	Name        string `json:"name"`
	Creator     string `json:"creator"`
	AppName     string `json:"app_name"`
	ChartLayout string `json:"chart_layout"`
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

type DashboardChartRelation struct {
	Entity
	DashboardId uint32 `json:"dashboard_id"`
	ChartId     uint32 `json:"chart_id"`
}

func (DashboardChartRelation) TableName() string {
	return "dashboard_chart_relation"
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
