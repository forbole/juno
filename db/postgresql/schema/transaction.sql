CREATE TABLE transaction
(
    hash         TEXT    NOT NULL UNIQUE PRIMARY KEY,
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
    gas_wanted   INTEGER          DEFAULT 0,
    gas_used     INTEGER          DEFAULT 0,
    raw_log      TEXT,
    logs         JSONB
);
