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

package pubsub

import (
	natsd "github.com/nats-io/nats-server/v2/server"
	"time"

	"github.com/jakexks/cert-manager-lifecycle-events/pkg/controller"
)

func Start(ctx *controller.Context) error {
	server, err := natsd.NewServer(&natsd.Options{
		ServerName: "cert-manager-lifecycle-events",
	})
	if err != nil {
		ctx.Log.Error(err, "couldn't create NATS server")
		return err
	}
	ctx.NatsServer = server
	go func() {
		if err := natsd.Run(server); err != nil {
			panic(err)
		}
	}()
	for !server.Running() {
		ctx.Log.Info("waiting for nats server to start...")
		time.Sleep(time.Second)
	}
	return nil
}
