CREATE TABLE transaction
(
    timestamp  TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    gas_wanted INTEGER                              DEFAULT 0,
    gas_used   INTEGER                              DEFAULT 0,
    height     BIGINT                      NOT NULL REFERENCES block (height),
    hash       TEXT                        NOT NULL UNIQUE PRIMARY KEY,
    messages   JSONB                       NOT NULL DEFAULT '[]'::JSONB,
    fee        JSONB                       NOT NULL DEFAULT '{}'::JSONB,
    signatures JSONB                       NOT NULL DEFAULT '[]'::JSONB,
    raw_log    TEXT,
    success    BOOLEAN                     NOT NULL DEFAULT true,
    memo       TEXT
);
