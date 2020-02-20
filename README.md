# Desmos Parser

My first contribute to Desmos <3 (alpha, alpha, alpha version)

## Installation

```bash
git clone git@github.com:angelorc/desmos-parser.git
cd desmos-parser
make install
```

## Config
config.toml
```bash
rpc_node = "http://lcd.morpheus.desmos.network:26657"
client_node = "http://rpc.morpheus.desmos.network:1317"

[database]
uri = "mongodb://localhost:27017/desmos?replicaSet=replica01"
name = "desmos"
```

## Run Parser
```bash
desmos-parser config.toml
```