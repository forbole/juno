# PostgreSQL Setup with GraphQL
In order to properly setup this project to work with PostgreSQL and, at the same time, create a GraphQL endpoint, you need to perform some additional setup. 

## Install PostgreSQL
First of all, install PostgreSQL following the [official documentation](https://www.postgresql.org/download/).

## Create the Desmos database 
Once you have installed PostgreSQL, you will need to create the Desmos database. To do so, follow the below steps. 

1. Log into PostgreSQL with root access.
   ```bash
    sudo -u postgres psql
   ``` 
   
2. Configure PostgreSQL to make is accessible by your normal users. Change your_username with your actual user already created on your Ubuntu system.
   ```postgresql
   CREATE ROLE <your_username> WITH SUPERUSER LOGIN ENCRYPTED PASSWORD 'your_password';
   ``` 
   
3. Exit PostgreSQL
   ```bash
   \q
   ```
   
4. Create the Desmos database and set yor user to be the owner. 
   ```bash
   createdb desmos -O <your-username>
   ```
   
5. Log into the Desmos database. 
   ```bash
   psql desmos
   ```
   
6. Create all the required tables. 
   ```postgresql
   CREATE TABLE validator (
       id SERIAL PRIMARY KEY,
       address character varying(40) NOT NULL UNIQUE,
       consensus_pubkey character varying(83) NOT NULL UNIQUE
   );
   
   CREATE TABLE pre_commit (
       id SERIAL PRIMARY KEY,
       height integer NOT NULL,
       round integer NOT NULL,
       validator_address character varying(40) NOT NULL REFERENCES validator(address),
       timestamp timestamp without time zone NOT NULL,
       voting_power integer NOT NULL,
       proposer_priority integer NOT NULL
   );
   
   CREATE TABLE block (
       id SERIAL PRIMARY KEY,
       height integer NOT NULL UNIQUE,
       hash character varying(64) NOT NULL UNIQUE,
       num_txs integer DEFAULT 0,
       total_gas integer DEFAULT 0,
       proposer_address character varying(40) NOT NULL REFERENCES validator(address),
       pre_commits integer NOT NULL,
       timestamp timestamp without time zone NOT NULL
   );
   
   CREATE TABLE transaction (
       id SERIAL PRIMARY KEY,
       timestamp timestamp without time zone NOT NULL,
       gas_wanted integer DEFAULT 0,
       gas_used integer DEFAULT 0,
       height integer NOT NULL REFERENCES block(height),
       txhash character varying(64) NOT NULL UNIQUE,
       events jsonb DEFAULT '[]'::jsonb,
       messages jsonb NOT NULL DEFAULT '[]'::jsonb,
       fee jsonb NOT NULL DEFAULT '{}'::jsonb,
       signatures jsonb NOT NULL DEFAULT '[]'::jsonb,
       memo character varying(256)
   );
   ``` 
   
7. Exit PostgreSQL. 
   ```bash
   \q
   ```
   
Now that the database has been created, you can try running the parser using the following configuration: 

```toml
rpc_node = "http://rpc.morpheus.desmos.network:26657"
client_node = "http://lcd.morpheus.desmos.network:1317"

[database]
name = "desmos"
host = "localhost"
port = 5432
user = "your_username"
password = "your_password"
```

## Setup the GraphQL APIs with Hasura
In order to easily setup the GraphQL APIs, we're going to use [Hasura](https://hasura.io/). This project will allow you to run a Docker container which exposes the GraphQL APIs allowing you to perform custom queries without much effort.  

### Setup the database
The first thing we need to do is setup the Desmos database so that Hasura can connect properly. To do so, follow the below steps.   

1. Log into the Desmos database
   ```bash 
   psql desmos
   ```

2. Create all the roles and permissions for Hasura to work
    ```postgresql
   -- We will create a separate user and grant permissions on hasura-specific
   -- schemas and information_schema and pg_catalog
   -- These permissions/grants are required for Hasura to work properly.
   
   -- create a separate user for hasura
   CREATE USER hasurauser WITH PASSWORD 'hasurauser';
   
   -- create pgcrypto extension, required for UUID
   CREATE EXTENSION IF NOT EXISTS pgcrypto;
   
   -- create the schemas required by the hasura system
   -- NOTE: If you are starting from scratch: drop the below schemas first, if they exist.
   CREATE SCHEMA IF NOT EXISTS hdb_catalog;
   CREATE SCHEMA IF NOT EXISTS hdb_views;
   
   -- make the user an owner of system schemas
   ALTER SCHEMA hdb_catalog OWNER TO hasurauser;
   ALTER SCHEMA hdb_views OWNER TO hasurauser;
   
   -- grant select permissions on information_schema and pg_catalog. This is
   -- required for hasura to query list of available tables
   GRANT SELECT ON ALL TABLES IN SCHEMA information_schema TO hasurauser;
   GRANT SELECT ON ALL TABLES IN SCHEMA pg_catalog TO hasurauser;
   
   -- Below permissions are optional. This is dependent on what access to your
   -- tables/schemas - you want give to hasura. If you want expose the public
   -- schema for GraphQL query then give permissions on public schema to the
   -- hasura user.
   -- Be careful to use these in your production db. Consult the postgres manual or
   -- your DBA and give appropriate permissions.
   
   -- grant all privileges on all tables in the public schema. This can be customised:
   -- For example, if you only want to use GraphQL regular queries and not mutations,
   -- then you can set: GRANT SELECT ON ALL TABLES...
   GRANT USAGE ON SCHEMA public TO hasurauser;
   GRANT ALL ON ALL TABLES IN SCHEMA public TO hasurauser;
   GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO hasurauser;
   
   -- Similarly add this for other schemas, if you have any.
   -- GRANT USAGE ON SCHEMA <schema-name> TO hasurauser;
   -- GRANT ALL ON ALL TABLES IN SCHEMA <schema-name> TO hasurauser;
   -- GRANT ALL ON ALL SEQUENCES IN SCHEMA <schema-name> TO hasurauser; 
   ```

## Get Hasura
The next thing we need to do is getting Hasura. To do so we need to perform the below steps.

1. Get the Hasura Docker script by executing the following command.  
   ```bash
   wget https://raw.githubusercontent.com/hasura/graphql-engine/stable/install-manifests/docker-run/docker-run.sh
   ```

2. Configure the script by editing its content so that it looks like the following.  
   **Note**. Replace `your_username` and `your_password` with the database username and password you set during the database creation.   
   ```bash
   #! /bin/bash
   docker run -d --net=host \
          -e HASURA_GRAPHQL_DATABASE_URL=postgres://your_username:your_password@localhost:5432/desmos \
          -e HASURA_GRAPHQL_ENABLE_CONSOLE=true \
          hasura/graphql-engine:v1.1.0
   ```

3. Make the script executable.
   ```bash
   chmod +x ./docker-run.sh
   ``` 
   
4. Start the script. 
   ```bash
   ./docker-run.sh
   ```
   
If everything works out, you should be able to see the local GraphQL APIs explorer by browsing [localhost:8080](http://localhost:8080). 

## Setup Hasura
Once Hasura is properly running, you need to perform the latest configuration to make sure it reads the proper data. 

To do so, log into Hasura and click on the _Data_ section on the top bar: 

[![](.img/hasura_data_screen.png)](http://localhost:8080/console/data/schema/public)

Now, from the _Untracked tables or views_ select the tables you want to track: 

![](.img/hasura_track_views.png)

Once you have selected the views to track you also need to select the foreign keys as well: 

![](.img/hasura_track_keys.png)

Once you have done so, by going inside the _GraphiQL_ section you will be able to compose your query using the left side panel, run it using the _Play_ button and seeing the result on the right side panel: 

![](.img/hasura_result.png) 
