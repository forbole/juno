CREATE TABLE block
(
    height           BIGINT PRIMARY KEY,
    hash             TEXT                        NOT NULL UNIQUE,
    num_txs          INTEGER DEFAULT 0,
    total_gas        INTEGER DEFAULT 0,
    proposer_address TEXT                        NOT NULL REFERENCES validator (consensus_address),
    timestamp        TIMESTAMP WITHOUT TIME ZONE NOT NULL
);
