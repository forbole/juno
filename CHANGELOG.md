## v3.0.0
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
