# nuchal
**nuchal**, *new⋅cull, nu⋅chal, /ˈn(y)o͞ok(ə)l/,* is an application for evaluating and executing cryptocurrency trades 
from based statistical pattern recognition on technical analysis using multiple timeframes. 

The goal of **nuchal** was to create a trading extension capable of simulating trades from historical rates, executing 
live trades with optimized limit orders, and providing detailed account position reports.

## Configuration
A [Coinbase Pro][1] account, a working installation of [GO][2], and a running instance of [Docker][3] **are required**.
### Coinbase
```shell
# Create a Coinbase Pro API and export the values
export COINBASE_PRO_KEY="your_coinbase_pro_api_key"
export COINBASE_PRO_PASSPHRASE="your_coinbase_pro_api_passphrase"
export COINBASE_PRO_SECRET="your_coinbase_pro_api_secret"
```

### GO
```shell
# Add the go bin directory to your system path
export PATH=${PATH}:/Users/${USER}/go/bin

# Download and install nuchal 
go get github.com/nelsw/nuchal
```

### Docker
```shell
# Start the nuchal docker composition (database)
docker compose -p nuchal -f build/docker-compose.yml up -d

# To power down the database, use the following command
docker compose -f build/docker-compose.yml down
```

### Products & Patterns
Products AKA cryptocurrencies are from the Coinbase API, there is no configuration required to get available products. 
Patterns are what define the parameters for making systematic trading decisions. If no pattern configuration is present, 
nuchal will create a "default" pattern for each product available. To config product patterns, create a `config.yml` 
similar to the following example and add it to the base project directory of nuchal.

```yaml
# Coinbase Pro configuration 
# with maker and taker fees
cbp:
  key:
  pass:
  secret:
  fees:
    maker:
    taker:

# define product patterns here
patterns:
  - id: SKL-USD
    delta: .01
  - id: NU-USD
    gain: .0273
  - id: OMG-USD
    loss: .0473
  - id: TRB-USD
    size: 1.25

# you can also define a period of time related to the 
# command this could be when start and end command
# execution or a range of data to simulate
period:
  alpha: 2021-06-02T08:00:00+00:00
  omega: 2022-06-03T22:00:00+00:00
```

## Commands

### report
```shell
# Prints USD, Cryptocurrency, and value of the configured Coinbase Pro Account.
nuchal report
```

### trade
```shell
# Trade, that is buy and sell configured products.
nuchal trade

# Hold the available balance for all configured products.	
nuchal trade --hold

# Sell the available balance for all configured products.
nuchal trade --sell

# Exit the available balance for all configured products.
nuchal trade --exit

```

### sim
```shell
# Prints a simulation result report and serves a local website for graphs of said simulation results.
nuchal sim
```

# License
This code is Copyright Connor Ross Van Elswyk and licensed under Apache License 2.0

[1]: https://pro.coinbase.com
[2]: https://golang.org/
[3]: https://www.docker.com/