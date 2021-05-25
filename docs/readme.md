# nuchal
In marine biology, **nuchal**, *new⋅cull, nu⋅chal, /ˈn(y)o͞ok(ə)l/,* is defined as 
> *" ... an exaggerated (craniofacial) trait, responsible for defining success among a species."*

Like its namesake, **nuchal** defines successful opportunities and creates trades based on these opportunities. 

At its core, **nuchal** is a project for automating cryptocurrency trades on Coinbase Pro. **nuchal** creates mainly 
profitable results when configured properly and run during specific market stages, namely the markup and distribution 
stage. However, **nuchal** is not perfect - and markets can stay irrational longer than we can stay solvent. That said, 
**nuchal** GUARANTEES NOTHING - USE AT YOUR OWN RISK.

## Overview
The **goals** of this project are to:
- recognize trade opportunities based on statistical pattern analysis
- automate market and limit trading based on pattern recognition
- simulate automation results using historical ticker data
- provide a streaming summary of portfolio positions

### Requirements
- a [Coinbase Pro][1] account
- a working installation of [GO][2]
- a running instance of [Docker][3]

### Configuration

#### Coinbase
```shell
# Create a Coinbase Pro API and export the values
export COINBASE_PRO_KEY="your_coinbase_pro_api_key"
export COINBASE_PRO_PASSPHRASE="your_coinbase_pro_api_passphrase"
export COINBASE_PRO_SECRET="your_coinbase_pro_api_secret"
```

#### GO
```shell
# Add the go bin directory to your system path
export PATH=${PATH}:/Users/${USER}/go/bin

# Download and install nuchal 
go get github.com/nelsw/nuchal
```

#### Docker
```shell
# Start the nuchal docker composition (database)
docker compose -p nuchal -f build/docker-compose.yml up -d

# To power down the database, use the following command
docker compose -f build/docker-compose.yml down
```

### Use
nuchal can simulate trades from historical rates, execute live trades with limit orders, and report on positions.

The scope of cryptocurrency "products" included in these commands can be modified by enabling or disabling a product
"pattern" in `pkg/config/patterns.json`. Product patterns also define the criteria used to recognize opportunities.   

Currently, nuchal runs off of a single "Tweezer Bottom" pattern to recognize trading opportunities. When the pattern is
recognized, the product is purchased at the market price. Once the order is complete, nuchal watches the product ticker 
and waits to place a stop entry limit order until the goal price or better is reached.

#### Trade
```shell
# 💎👐🏻
nuchal trade
```

#### Simulate
```shell
# Run simulation and print results to console.
nuchal sim

# Run simulation, print results to console, and serve charts to localhost.
nuchal sim --serve
```

#### Report
```shell
# Print report report stats.
nuchal report

# Print report report stats, every minute.
nuchal report --recurring

# Print report report stats, and place limit orders to hold the full balance.
nuchal report --force-holds
```

#### Multiple Users
Place a `.json` file in the `pkg/config` directory:
```json
{
  "users": [
    {
      "name": "your_name",
      "key": "your_coinbase_pro_api_key",
      "passphrase": "your_coinbase_pro_api_passphrase",
      "secret": "your_coinbase_pro_api_secret",
      "enable": true
    },
    {
      "name": "their_name",
      "key": "their_coinbase_pro_api_key",
      "passphrase": "their_coinbase_pro_api_passphrase",
      "secret": "their_coinbase_pro_api_secret",
      "enable": false
    }
  ]
}
```

# License
This code is Copyright Connor Ross Van Elswyk and licensed under Apache License 2.0

[1]: https://pro.coinbase.com
[2]: https://golang.org/
[3]: https://www.docker.com/