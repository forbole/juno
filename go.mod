module github.com/desmos-labs/juno

go 1.13

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/cosmos/cosmos-sdk v0.40.0-rc7
	github.com/go-co-op/gocron v0.3.3
	github.com/lib/pq v1.9.0
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.20.0
	github.com/spf13/cobra v1.1.1
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.6.1
	github.com/tendermint/tendermint v0.34.1
	go.mongodb.org/mongo-driver v1.3.0
	google.golang.org/grpc v1.33.2
	gopkg.in/yaml.v2 v2.4.0
)

replace google.golang.org/grpc => google.golang.org/grpc v1.33.2

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.2-alpha.regen.4
