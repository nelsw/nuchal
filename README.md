# nuchal [![Go Report Card][5]][4] [![PkgGoDev][7]][6]

Evaluates & executes high frequency cryptocurrency trades from configurable trend alignment patterns.

## Installation
A [Coinbase Pro][1] account, a working installation of [GO][2], and a running instance of [Docker][3] **are required**.
```shell
# Download and install the latest version of nuchal 
go install github.com/nelsw/nuchal@latest

# Add the go bin directory to your path
export PATH=${PATH}:/Users/${USER}/go/bin

# Confirm successful installation
nuchal
```

## Configuration

### Currency
**nuchal** supports Fiat (USD) -> Crypto and Crypto -> Fiat (USD) trading only.

### Products
**nuchal** caches all available USD cryptocurrency products from Coinbase on startup.
```json
{
  "id": "BTC-USD",
  "base_currency": "BTC",
  "quote_currency": "USD",
  "base_min_size": "0.0001",
  "base_max_size": "280",
  "quote_increment": "0.01"
}
```

### Patterns
**nuchal** supports a single "Tweezer Bottom" trend alignment pattern to recognize and define opportunities.
```json
{
 "id": "BTC-USD",
 "gain": 0.0195,
 "loss": 0.495,
 "size": 1,
 "delta": 0.001
}
```

### Sources
Not all commands work in *Sandbox* mode, *Production* mode requires configuration at least one **source**.

#### env
Execute the following if you're not interested in pattern configuration on a product level.
```shell
# Minimum requirement for running in Production mode: key, pass, secret.
export COINBASE_PRO_KEY="your_coinbase_pro_api_key"
export COINBASE_PRO_PASSPHRASE="your_coinbase_pro_api_passphrase"
export COINBASE_PRO_SECRET="your_coinbase_pro_api_secret"

# Coinbase Pro trading information for precise trade results.
export COINBASE_PRO_MAKER_FEE="your_coinbase_pro_api_maker_fee"
export COINBASE_PRO_TAKER_FEE="your_coinbase_pro_api_taker_fee"

# A time frame for the command or command data.
export PERIOD_ALPHA="2021-06-02T08:00:00+00:00"
export PERIOD_OMEGA="2021-06-03T22:00:00+00:00"
export PERIOD_DURATION="24h00m00s"
```

#### yml 
To support product based pattern definitions, create a `nuchal.yml` configuration file like the following.
**nuchal** will look in the home and desktop directories for this file. You may also define config location on the cli.
```yaml
# Minimum requirement for running in Production mode: key, pass, secret.
cbp:
  key:
  pass:
  secret:
  # Coinbase Pro trading information for precise trade results.
  fees:
    maker:
    taker:

# Product selection and pattern criteria.
patterns:
  - id: SKL-USD
    delta: .01
  - id: NU-USD
    gain: .0273
  - id: OMG-USD
    loss: .0473
  - id: TRB-USD
    size: 1.25

# A time frame for the command or command data.
period:
    alpha: 2021-06-02T00:00:00+00:00
    omega: 2021-06-03T23:59:59+00:00
    duration: 24h0m0s
```

#### cli
```shell
# Displays options for configuring global pattern criteria, USD selections, and configuration file location.
nuchal --help
```

## Commands
**nuchal** has three (3) main functions:
1. report
2. trade
3. simulate


### report
Provides a summary of your available currencies, balances, holds, and status of open trading positions.
```shell
# Prints USD, Cryptocurrency, and total value of the configured Coinbase Pro account.
# Also prints position and trading information, namely size, value, balance and holds.
nuchal report -c /Users/${USER}/config.yml
```
![report example][10]

### sim
Evaluates product & pattern configuration through a mock trading session and interactive chart results.
```shell
# Prints a simulation result report and serves a local website to host graphical report analysis.
nuchal sim

# Prints a simulation result report with a positive net gain.
nuchal sim -t --no-losers

# Prints a simulation result report with a positive net gain and zero trading positions. 
nuchal sim -w --winners-only
```
![sim example][12]
![chart example][14]

### trade
Polls ticker data and executes buy & sell orders when conditions match product & pattern configuration.
```shell
# Trade buys & sells products at prices or at times that meet or exceed pattern criteria, for a specified duration.
nuchal trade  --usd XLM,TRB,SKL,STORJ

# Hold creates a limit entry order at the goal price for every active trading position in your available balance.
nuchal trade --hold

# Sell all available positions (active trades) at prices or at times that meet or exceed pattern criteria.
nuchal trade --sell

# Sell all available positions (active trades) at the current market price. Will not sell holds.
nuchal trade --exit

# Drop will cancel every hold order, allowing the resulting products to be sold or converted.
nuchal trade --drop

# Sells everything at market price.
nuchal trade --eject
```
![trade example][11]

# Thanks
**nuchal** is built largely on [a Go client for CoinBase Pro][8] formerly known as gdax, thank you [preichenberger][9].

**nuchal** charts made possible by [go-echarts][16] 

# Questions

> "How do I exit nuchal?"
```shell
exit
```

> "How do stop the docker composition?"
```shell
# To stop container orchestration
docker compose -f /Users/${USER}/go/src/github.com/nelsw/nuchal/build/docker-compose.yml down
```

> "What does nuchal mean?" 

In marine biology, **nuchal**, *new⋅cull, nu⋅chal, /ˈn(y)o͞ok(ə)l/,* is defined as *"an exaggerated (craniofacial) trait, 
responsible for defining the natural order among a species."*

# License
This code is Copyright Connor Ross Van Elswyk and licensed under Apache License 2.0

[1]: https://pro.coinbase.com
[2]: https://golang.org/
[3]: https://www.docker.com/
[4]: https://goreportcard.com/report/github.com/nelsw/nuchal
[5]: https://goreportcard.com/badge/github.com/nelsw/nuchal
[6]: https://pkg.go.dev/mod/github.com/nelsw/nuchal
[7]: https://pkg.go.dev/badge/mod/github.com/nelsw/nuchal
[8]: https://github.com/preichenberger/go-coinbasepro
[9]: https://github.com/preichenberger
[10]: .github/report.png?raw=true
[11]: .github/trade.png?raw=true
[12]: .github/sim.png?raw=true
[13]: https://www.investopedia.com/articles/active-trading/040714/tweezers-provide-precision-trend-traders.asp
[14]: .github/charts.png?raw=true
[16]: https://github.com/go-echarts/go-echarts