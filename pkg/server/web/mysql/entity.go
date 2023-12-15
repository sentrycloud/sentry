package mysql

import (
	"github.com/sentrycloud/sentry/pkg/dbmodel"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/protocol"
	"net/http"
)

func queryAllEntities(w http.ResponseWriter, entities interface{}) {
	err := dbmodel.QueryAllEntity(&entities)
	if err != nil {
		protocol.WriteQueryResp(w, http.StatusInternalServerError, 1, "db query failed", nil)
	} else {
		protocol.WriteQueryResp(w, http.StatusOK, 0, "ok", entities)
	}
}

func modifyEntity(w http.ResponseWriter, r *http.Request, modifyFunc func(interface{}) error, entity interface{}) {
	err := protocol.Json.NewDecoder(r.Body).Decode(entity)
	if err != nil {
		newlog.Error("json decode failed: %v", err)
		protocol.WriteQueryResp(w, http.StatusInternalServerError, 2, "json decoder failed", nil)
		return
	}

	err = modifyFunc(entity)
	if err != nil {
		newlog.Error("db modify failed: %v", err)
		protocol.WriteQueryResp(w, http.StatusInternalServerError, 3, "db modify failed", nil)
		return
	}

	protocol.WriteQueryResp(w, http.StatusOK, 0, "ok", entity)
}
