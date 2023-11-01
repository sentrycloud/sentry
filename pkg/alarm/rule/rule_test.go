package rule

import (
	"fmt"
	"math"
	"testing"
	"time"
)

func TestMinMaxFloat64(t *testing.T) {
	fmt.Printf("min float64: %.10e\n", -math.MaxFloat64)
	fmt.Printf("max float64: %.10e\n", math.MaxFloat64)
	fmt.Printf("smallest nonzero float64: %.10e\n", math.SmallestNonzeroFloat64)
	x := 0.0
	z := 4.0 / x

	if z > math.MaxFloat64 {
		fmt.Println("inf > math.MaxFloat64")
	} else {
		fmt.Println("inf < math.MaxFloat64")
	}
}

func TestBuildAlarmMessage(t *testing.T) {
	var r = AlarmRule{}
	r.Message = "{time} {metric} {tags} {tag.ip} 当前值已经达到 {value}"

	metric := "test.metric"
	tags := make(map[string]string)
	tags["ip"] = "127.0.0.1"

	ts := time.Now().Unix()
	value := 120000.345

	message := r.buildAlarmMessage(metric, tags, ts, value)
	t.Log(message)
}
