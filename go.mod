module github.com/desmos-labs/juno

go 1.13

require (
	github.com/cosmos/cosmos-sdk v0.42.4
	github.com/go-co-op/gocron v0.3.3
	github.com/gogo/protobuf v1.3.3
	github.com/lib/pq v1.9.0
	github.com/pelletier/go-toml v1.8.0
	github.com/rs/zerolog v1.20.0
	github.com/spf13/cobra v1.1.1
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	github.com/tendermint/tendermint v0.34.9
	google.golang.org/grpc v1.35.0
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
