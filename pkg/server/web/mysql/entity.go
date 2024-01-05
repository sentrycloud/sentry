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
		protocol.WriteQueryResp(w, protocol.CodeExecMySQLError, nil)
	} else {
		protocol.WriteQueryResp(w, protocol.CodeOK, entities)
	}
}

func modifyEntity(w http.ResponseWriter, r *http.Request, modifyFunc func(interface{}) error, entity interface{}) {
	err := protocol.DecodeRequest(r, entity)
	if err != nil {
		newlog.Error("json decode failed: %v", err)
		protocol.WriteQueryResp(w, protocol.CodeJsonDecodeError, nil)
		return
	}

	err = modifyFunc(entity)
	if err != nil {
		newlog.Error("db modify failed: %v", err)
		protocol.WriteQueryResp(w, protocol.CodeExecMySQLError, nil)
		return
	}

	protocol.WriteQueryResp(w, protocol.CodeOK, entity)
}
