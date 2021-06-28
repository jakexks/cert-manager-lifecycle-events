module github.com/jakexks/cert-manager-lifecycle-events

go 1.16

require (
	github.com/go-logr/logr v0.4.0
	github.com/go-logr/zapr v0.4.0
	github.com/jetstack/cert-manager v1.4.0
	github.com/nats-io/nats-server/v2 v2.2.6
	github.com/nats-io/nats-streaming-server v0.22.0
	github.com/nats-io/nats.go v1.11.0 // indirect
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.7.0
	go.uber.org/zap v1.17.0
	k8s.io/client-go v0.21.0
	sigs.k8s.io/controller-runtime v0.9.0-beta.2
)
