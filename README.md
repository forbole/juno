# Juno
> This branch is intended to be used with Cosmos SDK `v0.39.x`.

[![Build Status](https://travis-ci.org/fissionlabsio/juno.svg?branch=master)](https://travis-ci.org/fissionlabsio/juno)
[![Go Report Card](https://goreportcard.com/badge/github.com/fissionlabsio/juno)](https://goreportcard.com/report/github.com/fissionlabsio/juno)

> Juno is a Cosmos Hub blockchain data aggregator and exporter that provides the ability for developers and clients to query for indexed chain data.

## Table of Contents
  - [Background](#background)
  - [Install](#install)
  - [Usage](#usage)
  - [Schemas](#schemas)
  - [GraphQL integration](#graphql-integration)
  - [Contributing](#contributing)
  - [License](#license)

## Background
This version of Juno is a fork of [FissionLabs's Juno](https://github.com/fissionlabsio/juno). 

The main reason behind the fork what to improve the original project by: 

1. allowing different databases types as data storage spaces;
2. creating a highly modular code that allows for easy customization.

We achieved the first objective by supporting both PostgreSQL and MongoDB. We also reviewed the code design by using a database interface so that you can implement whatever database backend you prefer most. 

On the other hand, to achieve a highly modular code, we implemented extension points through the `worker.RegisterBlockHandler`, `worker.RegisterTxHandler` and `worker.RegisterMsgHandler` methods. You can use those to extend the default working of the code (which simply parses and saves the data on the database) with whatever operation you want.    


## Install
Juno takes a simple configuration. It needs to only know about a database instance and a Tendermint RPC node.

To install the binary run `make install`.

**Note**: Requires [Go 1.13+](https://golang.org/dl/)

### Working with PostgreSQL
#### Config
```toml
rpc_node = "<rpc-ip/host>:<rpc-port>"
client_node = "<client-ip/host>:<client-port>"

[cosmos]
prefix = "desmos"
modules = []

[database]
type = "mongodb"

[database.config]
host = "<db-host>"
port = 5432
name = "<db-name>"
user = "<db-user>"
password = "<db-password>"
ssl_mode = "<ssl-mode>"
```

### Working with MongoDB
#### Config
```toml
rpc_node = "<rpc-ip/host>:<rpc-port>"
client_node = "<client-ip/host>:<client-port>"

[cosmos]
prefix = "desmos"
modules = []

[database]
type = "postgresql"

[database.config]
name = "<db-name>"
uri = "<mongodb-uri>"
```

## Usage
Juno internally runs a single worker that consumes from a single queue. The queue contains block heights to aggregate and export to a database. Juno will start a new block even listener where for each new block, it will enqueue the height. A worker listens for new heights and queries for various data related to the block height to persist. For each block height, Juno will persist the block, the validators that committed/signed the block, all the pre-commits for the block and the transactions in the block.

In addition, it will also sync missing blocks from `--start-height` to the latest known height.

```shell
$ juno parse /path/to/config.toml [flags]
```

## Schemas
The schema definitions are contained in the `schema/` directory. Note, these schemas are not necessarily optimal and are subject to change! However, feel free to fork this tool and expand upon the schemas as you see fit. Any tweaks will most likely require adjustments to the `database` wrapper.

## GraphQL integration
If you want to know how to run a GraphQL server that allows to expose the parsed data, please refer to the following guides: 

- [PostgreSQL setup with GraphQL](.docs/postgres-graphql-setup.md)

## Running as a service
If you want to run it as a service, you can follow [this guide](.docs/service-example.md).

## Contributing
Contributions are welcome! Please open an Issues or Pull Request for any changes.

## License
[CCC0 1.0 Universal](https://creativecommons.org/share-your-work/public-domain/cc0/)
