CREATE TABLE pre_commit
(
    id                SERIAL PRIMARY KEY,
    validator_address character varying(52)       NOT NULL REFERENCES validator (consensus_address),
    timestamp         timestamp without time zone NOT NULL,
    voting_power      integer                     NOT NULL,
    proposer_priority integer                     NOT NULL
);
