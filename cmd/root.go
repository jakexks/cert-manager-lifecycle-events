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

package cmd

import (
	"context"
	"fmt"
	"github.com/jakexks/cert-manager-lifecycle-events/pkg/eventer"
	"github.com/nats-io/nats.go"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-logr/zapr"
	"github.com/jakexks/cert-manager-lifecycle-events/pkg/controller"
	cmclient "github.com/jetstack/cert-manager/pkg/client/clientset/versioned"
	cminformers "github.com/jetstack/cert-manager/pkg/client/informers/externalversions"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/jakexks/cert-manager-lifecycle-events/pkg/pubsub"
)

var signalHandler = make(chan struct{})

func root(cmd *cobra.Command, args []string) {
	// Set up logger
	var log *zap.Logger
	if viper.GetBool("use-structured-logging") {
		l, err := zap.NewProduction()
		if err != nil {
			fmt.Fprintf(os.Stderr, "couldn't create zap logger: %s", err.Error())
			return
		}
		log = l

	} else {
		l, err := zap.NewDevelopment()
		if err != nil {
			fmt.Fprintf(os.Stderr, "couldn't create zap logger: %s", err.Error())
			return
		}
		log = l
	}
	logr := zapr.NewLogger(log)

	c, err := config.GetConfig()
	if err != nil {
		logr.Error(err, "couldn't find kube config")
	}
	cmClient, err := cmclient.NewForConfig(c)

	if err != nil {
		logr.Error(err, "couldn't create CM client")
		return
	}

	ctx := &controller.Context{
		Ctx: setupSignalHandler(),
		Log: logr.WithName("lifecycle-controller"),

		CmClient:          cmClient,
		CmInformerFactory: cminformers.NewSharedInformerFactory(cmClient, 10*time.Hour),
	}

	if err := pubsub.Start(ctx); err != nil {
		ctx.Log.Error(err, "couldn't run streaming server")
		return
	}

	natsClient, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		ctx.Log.Error(err, "couldn't construct NATS client")
		return
	}
	ctx.NatsClient = natsClient

	var urlTo *url.URL

	if len(viper.GetString("event-url")) > 0 {
		to, err := url.Parse(viper.GetString("event-url"))
		if err != nil {
			logr.Error(err, "--event-url is invalid")
			return
		}
		urlTo = to
	}


	testEventer := &eventer.EventSender{
		NatsClient: natsClient,
		Subject:    "io.cert-manager.>",
		To:         urlTo,
		Method: http.MethodPost,
		Log:        ctx.Log.WithName("eventer"),
	}

	go testEventer.Run(ctx.Ctx)

	if err := controller.Run(ctx); err != nil {
		ctx.Log.Error(err, "error in controller")
	}
}

// https://github.com/kubernetes-sigs/controller-runtime/blob/0f460129/pkg/manager/signals/signal.go
func setupSignalHandler() context.Context {
	close(signalHandler)
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cancel()
		<-c
		os.Exit(1)
	}()

	return ctx
}
