package http

import "net/http"

func queryTimeSeriesDataForLast(w http.ResponseWriter, r *http.Request) {
	queryTimeSeriesDataForRange(w, r)
}
