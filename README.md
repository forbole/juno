# Desmos Parser
My first contribute to Desmos <3 (alpha, alpha, alpha version)

## Requirements
### 1. Having Mongo installed.  
If you don't have Mongo, you can get it by following the official [download guide](https://docs.mongodb.com/manual/tutorial/install-mongodb-on-ubuntu/).  

### 2. Having a replica set created locally
If you don't know how to create one, just execute the following commands.

1. Stop the `mongod` service: 
   ```bash
   sudo systemctl stop mongod
   ```

2. Open the `mongod` service config file:  
   ```bash
   sudo nano /etc/mongod.conf
   ``` 
   
3. Uncomment the `#replication` line and paste the followings:  
   ```yaml 
   replication:
     replSetName: replica01 
   ```
   
4. Start `mongod` again:  
   ```bash
   sudo systemctl mongod start
   ```


## Installation
1. Clone this repository  
   ```bash
   git clone git@github.com:angelorc/desmos-parser.git
   cd desmos-parser
   make install
   ```
   
2. Install the binaries
   ```bash
   make install
   ```

## Configuration
The whole program configuration is done inside the `config.toml` file. 

```toml
rpc_node = "http://lcd.morpheus.desmos.network:26657"
client_node = "http://rpc.morpheus.desmos.network:1317"

[database]
uri = "mongodb://localhost:27017/desmos?replicaSet=replica01"
name = "desmos"
```

#### rpc_node
This variable identifies the RPC URL to which to connect to download the chain data. 

#### lcd_node
This variable tells which URL should be used to connect to the Cosmos Light Client Deamon. 

#### database
Contains the database configurations such as the uri and database name

## Running the parser
To run the parser simply execute the following command: 

```bash
desmosp parse config.toml
```
