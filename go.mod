module github.com/dapr-sandbox/components-go-sdk

go 1.19

replace github.com/dapr/dapr => github.com/mcandeia/dapr v0.0.0-20220928114040-c34d0c15f90b

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/mitchellh/mapstructure v1.5.1-0.20220423185008-bf980b35cac4 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sirupsen/logrus v1.9.0 // indirect
	golang.org/x/net v0.0.0-20220630215102-69896b714898 // indirect
	golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20220622171453-ea41d75dfa0f // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require (
	github.com/dapr/components-contrib v1.8.0-rc.1.0.20220924063709-b0d267317d47
	github.com/dapr/dapr v1.8.4-0.20220927170932-1c498253f24b
	github.com/dapr/kit v0.0.2
	github.com/google/uuid v1.3.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.8.0
	google.golang.org/grpc v1.48.0
	google.golang.org/protobuf v1.28.0
)
