module github.com/desmos-labs/juno

go 1.13

require (
	github.com/cosmos/cosmos-sdk v0.43.0
	github.com/cosmos/ibc-go v1.0.0
	github.com/ethereum/go-ethereum v1.10.3
	github.com/go-co-op/gocron v0.3.3
	github.com/gogo/protobuf v1.3.3
	github.com/gorilla/mux v1.8.0
	github.com/lib/pq v1.9.0
	github.com/pelletier/go-toml v1.9.3
	github.com/prometheus/client_golang v1.11.0
	github.com/prometheus/common v0.30.0 // indirect
	github.com/prometheus/procfs v0.7.1 // indirect
	github.com/rs/zerolog v1.23.0
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/tendermint/tendermint v0.34.12
	github.com/tharsis/ethermint v0.5.0
	golang.org/x/crypto v0.0.0-20210817164053-32db794688a5 // indirect
	golang.org/x/sys v0.0.0-20210823070655-63515b42dcdf // indirect
	google.golang.org/grpc v1.40.0
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1

replace github.com/tendermint/tendermint => github.com/forbole/tendermint v0.34.13-0.20210820072129-a2a4af55563d
