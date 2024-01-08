package mysql

import (
	"github.com/sentrycloud/sentry/pkg/dbmodel"
	"github.com/sentrycloud/sentry/pkg/protocol"
	"github.com/sentrycloud/sentry/pkg/server/monitor"
	"net/http"
	"time"
)

func HandleDashboard(w http.ResponseWriter, r *http.Request) {
	defer monitor.AddMonitorStats(time.Now(), "dashboard")

	var entity dbmodel.Dashboard
	switch r.Method {
	case "GET":
		var entities []dbmodel.Dashboard
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
