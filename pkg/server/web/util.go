package web

import (
	"errors"
	"fmt"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/protocol"
	"net/http"
	"strings"
	"time"
)

const (
	MaxQueryRange = 3600 * 24 * 5 // max time series data query range is 5 days
	MaxDownSample = 3600 * 24     // max down sample is 1 day
	MaxTagCount   = 16
)

const (
	CodeOK                 = 0
	CodeApiNotFound        = 1
	CodeJsonDecodeError    = 2
	CodeGetConnPoolError   = 3
	CodeExecSqlError       = 4
	CodeSplitTagsError     = 5
	CodeStarKeysError      = 6
	CodeMaxQueryRangeError = 7
	CodeAggregatorError    = 8
	CodeDownSampleError    = 9
	CodeMetricError        = 10
	CodeTagCountError      = 11
	CodeOrderError         = 12
)

var CodeMsg = map[int]string{
	CodeOK:                 "ok",
	CodeApiNotFound:        "api not found",
	CodeJsonDecodeError:    "json decode error",
	CodeGetConnPoolError:   "get conn pool error",
	CodeExecSqlError:       "exec SQL error",
	CodeSplitTagsError:     "split tags error",
	CodeStarKeysError:      "star keys error",
	CodeMaxQueryRangeError: "max query range error",
	CodeAggregatorError:    "aggregator error",
	CodeDownSampleError:    "down sample error",
	CodeMetricError:        "metric error",
	CodeTagCountError:      "too many tag count error",
	CodeOrderError:         "no such order",
}

func writeQueryResp(w http.ResponseWriter, status int, resp *protocol.QueryResp) {
	data, err := protocol.Json.Marshal(resp)
	if err != nil {
		newlog.Error("marsh query response failed")
		status = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(status)
	w.Write(data)
}

func WriteQueryResp(w http.ResponseWriter, status int, resp *protocol.QueryResp) {
	writeQueryResp(w, status, resp)
}

func splitTags(tags map[string]string) (map[string]string, map[string]string, error) {
	var starTags = make(map[string]string)
	var noStarTags = make(map[string]string)

	// split tags to query tags and non-query tags
	for key, value := range tags {
		if len(key) == 0 || len(value) == 0 {
			return nil, nil, errors.New("empty tag key or value")
		}

		if key == "*" && value == "*" {
			return nil, nil, errors.New("tag key and value are all *")
		}

		if strings.HasSuffix(value, "*") {
			starTags[key] = value[0 : len(value)-1]
		} else {
			noStarTags[key] = value
		}
	}

	return starTags, noStarTags, nil
}

// all metric and tag key will be included in “ to use the original name
func buildCurvesRequest(metric string, tags map[string]string, starTags map[string]string) (string, []string) {
	sqlFormat := "SELECT DISTINCT `%s` FROM `%s` WHERE %s;"

	var starKeys []string
	var starKeysCondition []string
	var prefixValueCondition []string
	for k, v := range starTags {
		starKeys = append(starKeys, k)
		starKeysCondition = append(starKeysCondition, "`"+k+"` IS NOT NULL")

		// prefix search
		if len(v) > 0 {
			prefix := fmt.Sprintf("`%s` like \"%s%%\"", k, v)
			prefixValueCondition = append(prefixValueCondition, prefix)
		}
	}
	selectKeys := strings.Join(starKeys, "`,`")

	var condition strings.Builder
	condition.WriteString(strings.Join(starKeysCondition, " AND "))
	for k, v := range tags {
		if strings.Contains(v, "'") {
			condition.WriteString(fmt.Sprintf(" AND `%s`=\"%s\"", k, v))
		} else {
			condition.WriteString(fmt.Sprintf(" AND `%s`='%s'", k, v))
		}
	}

	if len(prefixValueCondition) > 0 {
		condition.WriteString(" AND ")
		condition.WriteString(strings.Join(prefixValueCondition, " AND "))
	}

	return fmt.Sprintf(sqlFormat, selectKeys, metric, condition.String()), starKeys
}

func checkAndTransferTime(last int64, start *int64, end *int64) int {
	if last != 0 {
		// request with last field, so time range will be [now - last, now)
		if last > MaxQueryRange {
			return CodeMaxQueryRangeError
		}

		*end = time.Now().Unix()
		*start = *end - last
	} else {
		// request with start and end fields, so time range will be [start, end)
		if *start > *end {
			*start, *end = *end, *start // swap the start and end time when start > end
		}

		if *end-*start > MaxQueryRange {
			return CodeMaxQueryRangeError
		}
	}

	return CodeOK
}

func alignWithDownSample(downSample int64, start *int64, end *int64) int {
	if downSample == 0 || downSample > MaxDownSample {
		return CodeDownSampleError
	}

	*start = *start / downSample * downSample * 1000
	if *end%downSample == 0 {
		*end -= 1
	}
	*end *= 1000
	return CodeOK
}

func splitTagFilters(metric string, reqTags map[string]string) (string, map[string]string, map[string][]string, int) {
	if len(metric) == 0 {
		return "", nil, nil, CodeMetricError
	}

	if len(reqTags) > MaxTagCount {
		return "", nil, nil, CodeTagCountError
	}

	var starKey string
	tags := make(map[string]string)
	filters := make(map[string][]string)
	for k, v := range reqTags {
		if strings.Contains(k, "*") {
			return "", nil, nil, CodeStarKeysError
		}

		if strings.Contains(v, "*") {
			starKey = k
			continue
		}

		if strings.Contains(v, "||") {
			filters[k] = strings.Split(v, "||")
		} else {
			tags[k] = v
		}
	}

	return starKey, tags, filters, CodeOK
}

func transferTimeSeriesDataRequest(req *protocol.TimeSeriesDataRequest) int {
	code := checkAndTransferTime(req.Last, &req.Start, &req.End)
	if code != CodeOK {
		return code
	}

	code = alignWithDownSample(req.DownSample, &req.Start, &req.End)
	if code != CodeOK {
		return code
	}

	var err error
	req.Aggregator, err = protocol.CheckAggregator(req.Aggregator)
	if err != nil {
		return CodeAggregatorError
	}

	for _, m := range req.Metrics {
		var starKey string
		starKey, m.Tags, m.Filters, code = splitTagFilters(m.Metric, m.Tags)
		if code != CodeOK {
			return code
		}

		if len(starKey) > 0 {
			return CodeStarKeysError
		}
	}

	return CodeOK
}

func transferTopnRequest(req *protocol.TopNRequest) int {
	code := checkAndTransferTime(req.Last, &req.Start, &req.End)
	if code != CodeOK {
		return code
	}

	code = alignWithDownSample(req.DownSample, &req.Start, &req.End)
	if code != CodeOK {
		return code
	}

	var err error
	req.Aggregator, err = protocol.CheckAggregator(req.Aggregator)
	if err != nil {
		return CodeAggregatorError
	}

	req.Order, err = protocol.CheckOrder(req.Order)
	if err != nil {
		return CodeOrderError
	}

	req.Field, req.Tags, req.Filters, code = splitTagFilters(req.Metric, req.Tags)
	if code != CodeOK {
		return code
	}

	if len(req.Field) == 0 {
		return CodeStarKeysError
	}
	return CodeOK
}

func tagsToCondition(tags map[string]string) string {
	var condition strings.Builder
	firstKey := true
	for k, v := range tags {
		v = quoteValue(v)

		// if key starts with !=, use `key` != 'value` as condition
		op := "="
		if strings.HasPrefix(k, "!=") {
			delete(tags, k) // delete the not equal key so it will not appear in the returned tags map
			op = "!="
			k = k[2:]
		}

		if firstKey {
			firstKey = false
			condition.WriteString(fmt.Sprintf("`%s` %s %s", k, op, v))
		} else {
			condition.WriteString(fmt.Sprintf(" AND `%s` %s %s", k, op, v))
		}
	}

	return condition.String()
}

func filterTagsToCondition(filterTags map[string][]string) string {
	var condition strings.Builder
	firstKey := true
	for key, values := range filterTags {
		if firstKey {
			firstKey = false
			condition.WriteString("(")
		} else {
			condition.WriteString(" AND (")
		}

		// if key starts with !=, use `key` != 'value` as condition
		op := "="
		logicalOp := "OR"
		if strings.HasPrefix(key, "!=") {
			op = "!="
			logicalOp = "AND"
			key = key[2:]
		}

		firstFilter := true
		for _, v := range values {
			v = quoteValue(v)

			if firstFilter {
				firstFilter = false
				condition.WriteString(fmt.Sprintf("`%s` %s %s", key, op, v))
			} else {
				condition.WriteString(fmt.Sprintf(" %s `%s` %s %s", logicalOp, key, op, v))
			}
		}
		condition.WriteString(")")
	}

	return condition.String()
}

func quoteValue(v string) string {
	// if tag value contains single quote, use double quote to query, otherwise use single quote to query
	if strings.Contains(v, "'") {
		v = "\"" + v + "\""
	} else {
		v = "'" + v + "'"
	}
	return v
}

func buildRangeQuerySql(start int64, end int64, aggregator string, downSample int64, req *protocol.MetricReq) string {
	tagsCondition := tagsToCondition(req.Tags)
	if len(req.Filters) > 0 {
		tagsCondition += " AND " + filterTagsToCondition(req.Filters)
	}

	if len(tagsCondition) > 0 {
		tagsCondition = " AND " + tagsCondition
	}

	sqlFormat := "SELECT CAST(FIRST(_ts) as BIGINT),%s(_value) FROM `%s` WHERE _ts > %d AND _ts < %d %s INTERVAL(%ds)"
	return fmt.Sprintf(sqlFormat, aggregator, req.Metric, start, end, tagsCondition, downSample)
}

func buildTopnQuerySql(req *protocol.TopNRequest) string {
	tagsCondition := tagsToCondition(req.Tags)
	if len(req.Filters) > 0 {
		tagsCondition += " AND " + filterTagsToCondition(req.Filters)
	}

	if len(tagsCondition) > 0 {
		tagsCondition = " AND " + tagsCondition
	}

	sqlFormat := "SELECT `%s`,v FROM (SELECT `%s`,%s(_value) as v FROM `%s` WHERE _ts > %d AND _ts < %d %s AND `%s` IS NOT NULL GROUP BY `%s`) order by v %s limit %d;"
	return fmt.Sprintf(sqlFormat, req.Field, req.Field, req.Aggregator, req.Metric, req.Start, req.End, tagsCondition, req.Field, req.Field, req.Order, req.Limit)
}
