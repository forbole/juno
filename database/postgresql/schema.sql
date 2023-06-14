-------------------------------------------------------------------------------------------------------------------
--- Validators
-------------------------------------------------------------------------------------------------------------------

CREATE TABLE validator
(
    consensus_address TEXT NOT NULL PRIMARY KEY, /* Validator consensus address */
    consensus_pubkey  TEXT NOT NULL UNIQUE /* Validator consensus public key */
);

-------------------------------------------------------------------------------------------------------------------
--- Blocks
-------------------------------------------------------------------------------------------------------------------

CREATE TABLE block
(
    height           BIGINT UNIQUE PRIMARY KEY,
    hash             TEXT                        NOT NULL UNIQUE,
    num_txs          INTEGER DEFAULT 0,
    total_gas        BIGINT  DEFAULT 0,
    proposer_address TEXT REFERENCES validator (consensus_address),
    timestamp        TIMESTAMP WITHOUT TIME ZONE NOT NULL
);
CREATE INDEX block_height_index ON block (height);
CREATE INDEX block_hash_index ON block (hash);
CREATE INDEX block_proposer_address_index ON block (proposer_address);

CREATE TABLE pre_commit
(
    validator_address TEXT                        NOT NULL REFERENCES validator (consensus_address),
    height            BIGINT                      NOT NULL,
    timestamp         TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    voting_power      BIGINT                      NOT NULL,
    proposer_priority BIGINT                      NOT NULL,
    UNIQUE (validator_address, timestamp)
);
CREATE INDEX pre_commit_validator_address_index ON pre_commit (validator_address);
CREATE INDEX pre_commit_height_index ON pre_commit (height);

-------------------------------------------------------------------------------------------------------------------
--- Transactions
-------------------------------------------------------------------------------------------------------------------

CREATE TYPE COIN AS
(
    denom  TEXT,
    amount TEXT
);

CREATE TABLE transaction
(
    id           BIGSERIAL NOT NULL,

    hash         TEXT      NOT NULL,
    height       BIGINT    NOT NULL REFERENCES block (height),
    success      BOOLEAN   NOT NULL,
    memo         TEXT,

    /* Tx signing */
    signatures   TEXT[]    NOT NULL,

    /* Tx response */
    gas_wanted   BIGINT             DEFAULT 0,
    gas_used     BIGINT             DEFAULT 0,
    raw_log      TEXT,
    logs         JSONB,

    /* PSQL partition */
    partition_id BIGINT    NOT NULL DEFAULT 0,

    PRIMARY KEY (hash, partition_id)
) PARTITION BY LIST (partition_id);
CREATE INDEX transaction_hash_index ON transaction (hash);
CREATE INDEX transaction_height_index ON transaction (height);
CREATE INDEX transaction_partition_id_index ON transaction (partition_id);

CREATE TABLE transaction_fee
(
    id                       SERIAL NOT NULL PRIMARY KEY,

    transaction_hash         TEXT   NOT NULL,
    transaction_partition_id BIGINT NOT NULL,

    amount                   COIN[] NOT NULL DEFAULT '{}',
    gas_limit                TEXT   NOT NULL DEFAULT '',
    payer                    TEXT   NOT NULL DEFAULT '',
    granter                  TEXT   NOT NULL DEFAULT '',

    FOREIGN KEY (transaction_hash, transaction_partition_id) REFERENCES transaction (hash, partition_id)
);

CREATE TABLE transaction_signer_info
(
    id                       SERIAL NOT NULL PRIMARY KEY,

    transaction_hash         TEXT   NOT NULL,
    transaction_partition_id BIGINT NOT NULL,

    public_key               JSONB  NOT NULL,
    address                  TEXT   NOT NULL,
    mode_info                JSONB  NOT NULL,
    sequence                 TEXT   NOT NULL,

    FOREIGN KEY (transaction_hash, transaction_partition_id) REFERENCES transaction (hash, partition_id)
);

CREATE TABLE transaction_tip
(
    id                       SERIAL NOT NULL PRIMARY KEY,

    transaction_hash         TEXT   NOT NULL,
    transaction_partition_id BIGINT NOT NULL,

    tip                      JSONB  NOT NULL DEFAULT '{}',

    FOREIGN KEY (transaction_hash, transaction_partition_id) REFERENCES transaction (hash, partition_id)

);

-------------------------------------------------------------------------------------------------------------------
--- Messages
-------------------------------------------------------------------------------------------------------------------

CREATE TABLE message
(
    id                          SERIAL NOT NULL,

    transaction_hash            TEXT   NOT NULL,
    transaction_partition_id    BIGINT NOT NULL,

    index                       BIGINT NOT NULL,
    type                        TEXT   NOT NULL,
    value                       JSONB  NOT NULL,
    involved_accounts_addresses JSONB  NOT NULL DEFAULT '[]',

    /* PSQL partition */
    partition_id                BIGINT NOT NULL DEFAULT 0,

    FOREIGN KEY (transaction_hash, transaction_partition_id) REFERENCES transaction (hash, partition_id),
    CONSTRAINT unique_transaction_message UNIQUE (transaction_hash, transaction_partition_id, index, partition_id)
) PARTITION BY LIST (partition_id);
CREATE INDEX message_type_index ON message (type);
CREATE INDEX message_involved_accounts_index ON message USING GIN (involved_accounts_addresses);

/**
 * This function is used to find all the utils that involve any of the given addresses and have
 * type that is one of the specified types.
 */
CREATE FUNCTION messages_by_address(
    addresses TEXT[],
    types TEXT[],
    "limit" BIGINT = 100,
    "offset" BIGINT = 0)
    RETURNS SETOF message AS
$$
SELECT message.*
FROM message
         JOIN transaction ON message.transaction_hash = transaction.hash AND
                             message.transaction_partition_id = transaction.partition_id
WHERE (cardinality(types) = 0 OR type = ANY (types))
  AND involved_accounts_addresses ?| addresses
ORDER BY transaction.height DESC
LIMIT "limit" OFFSET "offset"
$$ LANGUAGE sql STABLE;