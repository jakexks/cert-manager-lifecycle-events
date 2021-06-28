/*
Copyright 2021 Jetstack Ltd.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package eventer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/jakexks/cert-manager-lifecycle-events/pkg/controller"
	"net/http"
	"net/url"

	"github.com/nats-io/nats.go"
)

type EventSender struct {
	NatsClient *nats.Conn
	Subject    string
	To         *url.URL
	Method     string

	Log logr.Logger
}

func (e *EventSender) Run(ctx context.Context) error {
	if e.To != nil {
		e.Log.Info("Will send events to", "URL", e.To.String())
	}
	sub, err := e.NatsClient.Subscribe(e.Subject, e.callback)
	if err != nil {
		return fmt.Errorf("coudn't subscribe to subject: %w", err)
	}
	<-ctx.Done()
	sub.Drain()
	if errors.Is(ctx.Err(), context.Canceled) {
		return nil
	}
	return ctx.Err()
}

func (e *EventSender) callback(m *nats.Msg) {
	event := new(controller.Message)
	json.Unmarshal(m.Data, event)
	e.Log.Info("received event",
		"subject", m.Subject,
		"operation", event.Operation,
		"for", event.CertSpec.DNSNames,
	)
	if e.To != nil {
		r, err := http.NewRequest(e.Method, e.To.String(), bytes.NewReader(m.Data))
		if err != nil {
			e.Log.Error(err, "couldn't construct http request")
			return
		}
		r.Header.Set("User-Agent", "cert-manager-lifecycle-events/0.1")
		resp, err := http.DefaultClient.Do(r)
		if err != nil {
			e.Log.Error(err, "couldn't send http request", "url", e.To.String())
			return
		}
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			e.Log.Error(err, "received a non OK http status", "url", e.To.String(), "code", resp.StatusCode)
		}
		resp.Body.Close()
	}
}
