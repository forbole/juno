CREATE TABLE pre_commit
(
    id                SERIAL PRIMARY KEY,
    validator_address TEXT                        NOT NULL REFERENCES validator (consensus_address),
    height            BIGINT                      NOT NULL,
    timestamp         TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    voting_power      INTEGER                     NOT NULL,
    proposer_priority INTEGER                     NOT NULL
);
