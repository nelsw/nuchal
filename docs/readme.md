# nuchal
**nuchal**, *new⋅cull, nu⋅chal, /ˈn(y)o͞ok(ə)l/,* - a program for trading cryptocurrency on Coinbase Pro for *1-n* users.

## Overview
The **goals** of this project are to:
- recognize trade opportunities based on statistical pattern analysis
- automate market and limit trading based on pattern recognition
- simulate automation results using historical ticker data
- provide a streaming summary of portfolio positions

### Setup
This project requires a [Coinbase Pro][1] API key and working installations of [Go][2], [git][3], and [Docker][4].

#### Configure
```shell
export COINBASE_PRO_KEY="your_coinbase_pro_api_key"
export COINBASE_PRO_PASSPHRASE="your_coinbase_pro_api_passphrase"
export COINBASE_PRO_SECRET="your_coinbase_pro_api_secret"
```

#### Build
```shell
export PATH=${PATH}:/Users/${USER}/go/bin
git clone https://github.com/nelsw/nuchal.git &&
cd nuchal &&
go mod download &&
go mod tidy &&
go build -o build/nuchal main.go &&
go install
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
# start the nuchal docker composition (database)
docker compose -p nuchal -f build/docker-compose.yml up -d
```
```shell
# run simulation and print results to console.
nuchal sim

# run simulation, print results to console, and serve charts to localhost.
nuchal sim --serve
```
```shell
# stop the nuchal docker composition (database)
docker compose -f build/docker-compose.yml down
```

#### Report
```shell
# print report report stats.
nuchal report

# print report report stats, every minute.
nuchal report --recurring

# print report report stats, and place limit orders to hold the full balance.
nuchal report --force-holds
```

#### Multiple Users
Place a `.json` file in the config directory:
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
[3]: https://git-scm.com/
[4]: https://www.docker.com/