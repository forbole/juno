CREATE TABLE block
(
    height           integer PRIMARY KEY,
    hash             character varying(64)       NOT NULL UNIQUE,
    num_txs          integer DEFAULT 0,
    total_gas        integer DEFAULT 0,
    proposer_address character varying(52)       NOT NULL REFERENCES validator (consensus_address),
    pre_commits      integer                     NOT NULL,
    timestamp        timestamp without time zone NOT NULL
);
