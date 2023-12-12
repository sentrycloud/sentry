package mysql

import (
	"github.com/sentrycloud/sentry/pkg/dbmodel"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/protocol"
	"net/http"
)

func HandleContact(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		queryContact(w, r)
	case "PUT":
		addContact(w, r)
	case "POST":
		updateContact(w, r)
	case "DELETE":
		deleteContact(w, r)
	default:
		protocol.MethodNotSupport(w, r)
	}
}

func queryContact(w http.ResponseWriter, r *http.Request) {
	contacts, err := dbmodel.QueryAllContacts()
	if err != nil {
		protocol.WriteQueryResp(w, http.StatusInternalServerError, 1, "db query failed", nil)
	} else {
		protocol.WriteQueryResp(w, http.StatusOK, 0, "ok", contacts)
	}
}

func addContact(w http.ResponseWriter, r *http.Request) {
	modifyContact(w, r, dbmodel.AddContact)
}

func updateContact(w http.ResponseWriter, r *http.Request) {
	modifyContact(w, r, dbmodel.UpdateContact)
}

func deleteContact(w http.ResponseWriter, r *http.Request) {
	modifyContact(w, r, dbmodel.DeleteContact)
}

func modifyContact(w http.ResponseWriter, r *http.Request, modifyFunc func(contact *dbmodel.AlarmContact) error) {
	var contact dbmodel.AlarmContact
	err := protocol.Json.NewDecoder(r.Body).Decode(&contact)
	if err != nil {
		newlog.Error("json decode failed: %v", err)
		protocol.WriteQueryResp(w, http.StatusInternalServerError, 2, "json decoder failed", nil)
		return
	}

	err = modifyFunc(&contact)
	if err != nil {
		newlog.Error("db modify failed: %v", err)
		protocol.WriteQueryResp(w, http.StatusInternalServerError, 3, "db modify failed", nil)
		return
	}

	protocol.WriteQueryResp(w, http.StatusOK, 0, "ok", contact)
}
