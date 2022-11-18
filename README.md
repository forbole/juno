<div align="center">
  <h1> Juno </h1>
</div>

![banner](.docs/.img/logo.png)

[![GitHub Workflow Status](https://img.shields.io/github/workflow/status/desmos-labs/juno/Tests)](https://github.com/saifullah619/juno/actions?query=workflow%3ATests)
[![Go Report Card](https://goreportcard.com/badge/github.com/saifullah619/juno)](https://goreportcard.com/report/github.com/saifullah619/juno)
![GitHub tag (latest SemVer)](https://img.shields.io/github/v/tag/forbole/juno)

> Juno is a Cosmos Hub blockchain data aggregator and exporter that provides the ability for developers and clients to query for indexed chain data.

## Table of Contents
  - [Background](#background)
  - [Usage](#usage)
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

## Compatibility table
Since the Cosmos SDK has evolved a lot, we have different versions of Juno available.

|               Cosmos SDK Version                |   Juno branch    |
|:-----------------------------------------------:|:----------------:|
|                    `v0.37.x`                    | `cosmos-v0.37.x` |
|                    `v0.38.x`                    | `cosmos-v0.38.x` |
|                    `v0.39.x`                    | `cosmos-v0.39.x` |
| Stargate <br> (`v0.40.x`, `v0.41.x`, `v0.42.x`) | `cosmos-v0.40.x` |
|         `v0.43.x`, `v0.44.x`, `v0.45.x`         | `cosmos-v0.44.x` |

## Usage
To know how to setup and run Juno, please refer to the [docs folder](.docs).

## Testing
If you want to test the code, you can do so by running

```shell
$ make test-unit
```

**Note**: Requires [Docker](https://docker.com).

This will:
1. Create a Docker container running a PostgreSQL database.
2. Run all the tests using that database as support.

## GraphQL integration
If you want to know how to run a GraphQL server that allows to expose the parsed data, please refer to the following guides: 

- [PostgreSQL setup with GraphQL](.docs/postgres-graphql-setup.md)

## Contributing
Contributions are welcome! Please open an Issues or Pull Request for any changes.

## License
[CCC0 1.0 Universal](https://creativecommons.org/share-your-work/public-domain/cc0/)
