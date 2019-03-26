# cortana

A Cosmos Hub data aggregator and exporter. For now, cortana remains simple in
that is processes blocks and relevant data (e.g. validators and pre-commits) and
stores them in a relational database allowing for unique and interesting queries
of the Cosmos Hub as a whole. Cortana spawned out of the simple interest in
queries like what is the average gas cost of a block?

__TODO__:

* Sync incoming new blocks from event feed
* Persist governance proposals and tallies
* Persist transactions

## Config

Cortana takes a simple configuration. It needs to only know about a Postgres
instance and a Tendermint RPC node.

Example:

```toml
node = "<rpc-ip/host>:<rpc-port>"

[database]
host = "<db-host>"
port = <db-port>
name = "<db-name>"
user = "<db-user>"
password = "<db-password>"
```

__Note__: The config will most likely need access to a lite/REST client as well
to aggregate and persist other data such as governance proposals.

## Usage

Cortana internally runs a single worker that consumes from a single queue. The
queue contains block heights to aggregate and export to Postgres. Initially, it
will export data starting from the latest block height it has stored until the
latest known height on the chain. Any failed job (block height) is re-enqueued.

Additionally, a `--sync-missing` flag can be provided to sync any failed or
missing data from a previous export.

For each block, cortana will persist the block, the validators that committed/signed
the block, and all the pre-commits for the block.

```shell
$ cortana --config=<path/to/config> [flags]
```

## Schemas

The schema definitions are contained in the `schema/` directory. Note, these
schemas are not necessarily optimal and are subject to change! However, feel
free to fork this tool and expand upon the schemas as you see fit. Any tweaks
will most likely require adjustments to the `database` wrapper.
