CREATE TABLE validator
(
    consensus_address TEXT NOT NULL PRIMARY KEY, /* Validator consensus address */
    consensus_pubkey  TEXT NOT NULL UNIQUE /* Validator consensus public key */
);

CREATE TABLE block
(
    height           BIGINT  UNIQUE PRIMARY KEY,
    hash             TEXT    NOT NULL UNIQUE,
    num_txs          INTEGER DEFAULT 0,
    total_gas        BIGINT  DEFAULT 0,
    proposer_address TEXT REFERENCES validator (consensus_address),
    timestamp        TIMESTAMP WITHOUT TIME ZONE NOT NULL
);
CREATE INDEX block_hash_index ON block (hash);
CREATE INDEX block_proposer_address_index ON block (proposer_address);

CREATE TABLE pre_commit
(
    validator_address TEXT                        NOT NULL REFERENCES validator (consensus_address),
    height            BIGINT                      NOT NULL,
    timestamp         TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    voting_power      BIGINT                      NOT NULL,
    proposer_priority INTEGER                     NOT NULL,
    UNIQUE (validator_address, timestamp)
);
CREATE INDEX pre_commit_validator_address_index ON pre_commit (validator_address);
CREATE INDEX pre_commit_height_index ON pre_commit (height);

CREATE TABLE transaction
(
    hash         TEXT    NOT NULL,
    height       BIGINT  NOT NULL REFERENCES block (height),
    success      BOOLEAN NOT NULL,

    /* Body */
    messages     JSONB   NOT NULL DEFAULT '[]'::JSONB,
    memo         TEXT,
    signatures   TEXT[]  NOT NULL,

    /* AuthInfo */
    signer_infos JSONB   NOT NULL DEFAULT '[]'::JSONB,
    fee          JSONB   NOT NULL DEFAULT '{}'::JSONB,

    /* Tx response */
    gas_wanted   BIGINT           DEFAULT 0,
    gas_used     BIGINT           DEFAULT 0,
    raw_log      TEXT,
    logs         JSONB,

    /* Psql partition */
    partition_id BIGINT NOT NULL PRIMARY KEY

)PARTITION BY LIST(partition_id);
CREATE INDEX transaction_hash_index ON transaction (hash);
CREATE INDEX transaction_height_index ON transaction (height);
CREATE INDEX transaction_partition_id_index ON transaction (partition_id);

CREATE TABLE message
(
    transaction_hash            TEXT   NOT NULL,
    index                       BIGINT NOT NULL,
    type                        TEXT   NOT NULL,
    value                       JSONB  NOT NULL,
    involved_accounts_addresses TEXT[] NULL,

    /* Psql partition */
    partition_id BIGINT REFERENCES transaction (partition_id)
)PARTITION BY LIST(partition_id);
CREATE INDEX message_transaction_hash_index ON message (transaction_hash);

CREATE TABLE pruning
(
    last_pruned_height BIGINT NOT NULL
)