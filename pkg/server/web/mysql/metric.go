package mysql

import (
	"github.com/sentrycloud/sentry/pkg/dbmodel"
	"github.com/sentrycloud/sentry/pkg/protocol"
	"net/http"
)

func HandleMetricWhiteList(w http.ResponseWriter, r *http.Request) {
	var entity dbmodel.MetricWhiteList
	switch r.Method {
	case "GET":
		var entities []dbmodel.MetricWhiteList
		queryAllEntities(w, entities)
	case "PUT":
		modifyEntity(w, r, dbmodel.AddEntity, &entity)
	case "POST":
		modifyEntity(w, r, dbmodel.UpdateEntity, &entity)
	case "DELETE":
		modifyEntity(w, r, dbmodel.DeleteEntity, &entity)
	default:
		protocol.MethodNotSupport(w)
	}
}
