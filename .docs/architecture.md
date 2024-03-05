# Architecture

Juno's architecture is quite straightforward. It operates by utilizing a `HeightQueue`, which is essentially a buffered
channel of block heights. This queue gets populated with consecutive block heights when the `start` command is
initiated. The heights added to this queue are determined by the parameters provided to the command, such as the `start`
and `end` height, in addition to the current chain height and the latest height stored in the database.

To populate the queue, a simple `for` loop is utilized, iterating from the calculated `start` height to the
computed `end` height. If no end height is specified, the parser will continuously enqueue new heights.

Simultaneously, there are `n` workers (the number can be specified in the configuration) constantly monitoring the
queue. These workers perform the following actions for each block height:

1. Query the details of the block with the given height from the chain.
2. Parse the transactions and messages within the block, along with other relevant details.
3. Store the block, transactions, and messages data in a PostgreSQL database.
4. Pass the parsed data to additional registered modules managed by the `Registrar` interface. These modules can then
   execute additional actions tailored to specific use cases for the chain. For
   instance, [Athena](https://github.com/desmos-labs/athena), which is a custom Juno implementation by Desmos, also
   stores data related to Desmos profiles, relationships, and other Desmos-specific elements.

![Architecture](./.img/architecture.png)