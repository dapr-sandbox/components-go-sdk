module github.com/dapr-sandbox/components-go-sdk

go 1.19

require (
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sirupsen/logrus v1.9.0 // indirect
	golang.org/x/net v0.0.0-20220630215102-69896b714898 // indirect
	golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20220622171453-ea41d75dfa0f // indirect
)

require (
	github.com/dapr/components-contrib v1.8.0-rc.1.0.20220901165827-19341e5a0ff4
	github.com/dapr/dapr v1.8.4-0.20220909163359-efaca389cc32
	github.com/dapr/kit v0.0.2
	google.golang.org/grpc v1.47.0
	google.golang.org/protobuf v1.28.0
)

replace github.com/dapr/dapr => github.com/mcandeia/dapr v0.0.0-20220913221641-0c6b9f5583c7
