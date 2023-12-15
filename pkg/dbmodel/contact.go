package dbmodel

type AlarmContact struct {
	Entity
	Name   string `json:"name"`
	Phone  string `json:"phone"`
	Mail   string `json:"mail"`
	Wechat string `json:"wechat"`
}

func (AlarmContact) TableName() string {
	return "alarm_contact"
}
