## v3.2.1
### Changes
- ([\#68](https://github.com/forbole/juno/pull/68)) Added `chain_id` label to prometheus to improve alert monitoring 

## v3.2.0
### Changes
- ([\#61](https://github.com/forbole/juno/pull/61)) Updated v3 migration code to handle database users names with a hyphen 
- ([\#59](https://github.com/forbole/juno/pull/59)) Added `parse transactios` command to re-fetch missing or incomplete transactions

## v3.1.1
### Changes
- Updated IBC to `v3.0.0`

## v3.1.0
### Changes
- Allow to return any `WritableConfig` when initializing the configuration

## v3.0.1
### Changes
- Updated IBC to `v2.2.0`

## v3.0.0
#### Note
Some changes included in this version are **breaking** due to the command renames. Extra precaution needs to be used when updating to this version.

### Migrating
To migrate to this version you can run the following command: 
```
juno migrate v3
```

### Changes 
#### CLI
- Renamed the `parse` command to `start`
- Renamed the `fix` command to `parse`

#### Database
- Store transactions and messages inside partitioned tables

### New features
#### CLI
- Added a `genesis-file` subcommand to the `parse` command that allows you to parse the genesis file only
