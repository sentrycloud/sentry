package profile

import (
	"fmt"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"net/http"
	_ "net/http/pprof"
	"time"
)

// StartProfileInBlockMode
// see profile information in http://ip:profilePort/debug/pprof/
// this function must be called in the last part of a main function, or in a goroutine
// cause this call will block the caller
func StartProfileInBlockMode(profilePort int) {
	err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", profilePort), nil)
	if err != nil {
		newlog.Error("profile listen failed: %v\n", err)
		for {
			time.Sleep(1 * time.Second)
		}
	}
}
