module github.com/amammay/gotoproduction

go 1.16

require (
	cloud.google.com/go v0.84.0
	cloud.google.com/go/firestore v1.5.0
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace v0.20.1
	github.com/Microsoft/go-winio v0.5.0 // indirect
	github.com/amammay/propagationgcp v0.0.3
	github.com/blendle/zapdriver v1.3.1
	github.com/containerd/containerd v1.5.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/gorilla/mux v1.8.0
	github.com/matryer/is v1.4.0
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/testcontainers/testcontainers-go v0.11.0
	go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux v0.20.0
	go.opentelemetry.io/otel v0.20.0
	go.opentelemetry.io/otel/sdk v0.20.0
	go.opentelemetry.io/otel/trace v0.20.0
	go.uber.org/atomic v1.8.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	go.uber.org/zap v1.17.0
	golang.org/x/net v0.0.0-20210504132125-bbd867fde50d // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	google.golang.org/grpc v1.38.0
)
