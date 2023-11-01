package http

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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(data)
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

// all metric and tag key will use “ to use the original name
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

func transferTimeSeriesDataRequest(req *protocol.TimeSeriesDataRequest) int {
	if req.Last != 0 {
		if req.Last > MaxQueryRange {
			return CodeMaxQueryRangeError
		}

		req.End = time.Now().Unix()
		req.Start = req.End - req.Last
	} else {
		if req.Start > req.End {
			req.Start, req.End = req.End, req.Start // swap the start and end time
		}

		if req.End-req.Start > MaxQueryRange {
			return CodeMaxQueryRangeError
		}
	}

	// alignment with down sample
	req.Start = req.Start / int64(req.DownSample) * int64(req.DownSample) * 1000
	if req.End%int64(req.DownSample) == 0 {
		req.End -= 1
	}
	req.End *= 1000

	req.Aggregator = strings.ToLower(req.Aggregator)
	if (req.Aggregator != "sum") && (req.Aggregator != "avg") && (req.Aggregator != "max") && (req.Aggregator != "min") {
		return CodeAggregatorError
	}

	if req.DownSample == 0 || req.DownSample > MaxDownSample {
		return CodeDownSampleError
	}

	for _, m := range req.Metrics {
		if len(m.Metric) == 0 {
			return CodeMetricError
		}

		if len(m.Tags) > MaxTagCount {
			return CodeTagCountError
		}

		tags := make(map[string]string)
		filters := make(map[string][]string)
		for k, v := range m.Tags {
			if strings.Contains(k, "*") || strings.Contains(v, "*") {
				return CodeStarKeysError
			}

			if strings.Contains(v, "||") {
				filters[k] = strings.Split(v, "||")
			} else {
				tags[k] = v
			}
		}

		if len(filters) > 0 {
			m.Tags = tags
			m.Filters = filters
		}
	}

	return CodeOK
}

func transferTopnRequest(req *protocol.TopNRequest) int {
	if req.Start > req.End {
		req.Start, req.End = req.End, req.Start // swap the start and end time
	}

	if req.End-req.Start > MaxQueryRange {
		return CodeMaxQueryRangeError
	}

	// alignment with down sample
	req.Start = req.Start / int64(req.DownSample) * int64(req.DownSample) * 1000
	if req.End%int64(req.DownSample) == 0 {
		req.End -= 1
	}
	req.End *= 1000

	req.Aggregator = strings.ToLower(req.Aggregator)
	if (req.Aggregator != "sum") && (req.Aggregator != "avg") && (req.Aggregator != "max") && (req.Aggregator != "min") {
		return CodeAggregatorError
	}

	if len(req.Order) == 0 {
		req.Order = "desc"
	}
	req.Order = strings.ToLower(req.Order)
	if (req.Order != "desc") && (req.Order != "asc") {
		return CodeOrderError
	}

	if req.DownSample == 0 || req.DownSample > MaxDownSample {
		return CodeDownSampleError
	}

	if len(req.Metric) == 0 {
		return CodeMetricError
	}

	if len(req.Tags) > MaxTagCount {
		return CodeTagCountError
	}

	tags := make(map[string]string)
	filters := make(map[string][]string)
	for k, v := range req.Tags {
		if v == "*" {
			if len(req.Field) == 0 {
				req.Field = k
				continue
			} else {
				return CodeStarKeysError
			}
		}

		if strings.Contains(v, "||") {
			filters[k] = strings.Split(v, "||")
		} else {
			tags[k] = v
		}
	}

	if len(req.Field) == 0 {
		return CodeStarKeysError
	}

	req.Tags = tags
	req.Filters = filters

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

func buildRangeQuerySql(start int64, end int64, aggregator string, downSample int, req *protocol.MetricReq) string {
	tagsCondition := tagsToCondition(req.Tags)
	if len(req.Filters) > 0 {
		tagsCondition += " AND " + filterTagsToCondition(req.Filters)
	}

	var sql string
	if len(tagsCondition) == 0 {
		sqlFormat := "SELECT CAST(FIRST(_ts) as BIGINT),%s(_value) FROM `%s` WHERE _ts > %d AND _ts < %d INTERVAL(%ds)"
		sql = fmt.Sprintf(sqlFormat, aggregator, req.Metric, start, end, downSample)
	} else {
		sqlFormat := "SELECT CAST(FIRST(_ts) as BIGINT),%s(_value) FROM `%s` WHERE _ts > %d AND _ts < %d AND %s INTERVAL(%ds)"
		sql = fmt.Sprintf(sqlFormat, aggregator, req.Metric, start, end, tagsCondition, downSample)
	}
	return sql
}

func buildTopnQuerySql(req *protocol.TopNRequest) string {
	tagsCondition := tagsToCondition(req.Tags)
	if len(req.Filters) > 0 {
		tagsCondition += " AND " + filterTagsToCondition(req.Filters)
	}

	if len(tagsCondition) > 0 {
		tagsCondition = " AND " + tagsCondition
	}

	sqlFormat := "SELECT `%s`,v FROM (SELECT `%s`,%s(_value) as v FROM `%s` WHERE _ts > %d AND _ts < %d %s AND `%s` IS NOT NULL GROUP BY `%s`)" +
		" order by v %s limit %d;"
	return fmt.Sprintf(sqlFormat, req.Field, req.Field, req.Aggregator, req.Metric, req.Start, req.End, tagsCondition, req.Field, req.Field, req.Order, req.Limit)
}
