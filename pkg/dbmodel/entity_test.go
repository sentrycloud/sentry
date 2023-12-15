package dbmodel

import (
	"fmt"
	"testing"
)

func TestGetField(t *testing.T) {
	var contact = AlarmContact{}
	fields := getJsonTags(&contact)
	fmt.Println(fields)

	var metric = MetricWhiteList{}
	fields = getJsonTags(&metric)
	fmt.Println(fields)
}
