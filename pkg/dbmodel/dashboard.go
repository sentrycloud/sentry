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
	Name        string `json:"name"`
	Type        int    `json:"type"`
	Aggregation string `json:"aggregation"`
	DownSample  string `json:"down_sample"`
	TopnLimit   int    `json:"topn_limit"`
}

func (Chart) TableName() string {
	return "chart"
}

type Line struct {
	Entity
	Name   string `json:"name"`
	Metric string `json:"metric"`
	Tags   string `json:"tags"`
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

type ChartLineRelation struct {
	Entity
	ChartId uint32 `json:"chart_id"`
	LineId  uint32 `json:"line_id"`
}

func (ChartLineRelation) TableName() string {
	return "chart_line_relation"
}
