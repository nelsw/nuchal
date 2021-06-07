# nuchal [![Go Report Card][5]][4] [![PkgGoDev][7]][6]

An application for evaluating and executing high frequency cryptocurrency trades using statistical pattern recognition. 

Effectively tests patterns against historical rates, executes trades with optimized limits, and prints detailed reports.

## Requirements
A [Coinbase Pro][1] account, a working installation of [GO][2], and a running instance of [Docker][3] **are required**.

## Installation

### GO
```shell
# Download and install nuchal 
go get -u github.com/nelsw/nuchal
```

### Docker
```shell
# Start the docker composition (postgres database)
docker compose -p nuchal -f build/docker-compose.yml up -d
```

## Configuration

Products are cryptocurrencies from the Coinbase API. There is no configuration for products. All products are available,
and while **nuchal** is designed to support multiple currencies, it is currently developed to support USD only.

Patterns are the criteria used to recognize trends and make critical trading decisions. Default criteria is overridable.
While **nuchal** is designed to support multiple patterns, it is currently developed to support only a "Tweezer Bottom".

Period is the window of time for command execution and data time frames. Default time frame is from 24hrs ago until now.

**nuchal** supports configuration from multiple sources and in the following order, sorted by highest precedence:
1. cli - command line interface
2. env - environment variables
3. yml - .yml configuration file

### cli
Start here.
```shell
# Displays options for configuring global product pattern criteria, USD selections, and configuration file location.
nuchal --help
```

### env
If you're just looking to get this up and running, env configuration is your friend.
```shell
# Create a Coinbase Pro API and export the values
export COINBASE_PRO_KEY="your_coinbase_pro_api_key"
export COINBASE_PRO_PASSPHRASE="your_coinbase_pro_api_passphrase"
export COINBASE_PRO_SECRET="your_coinbase_pro_api_secret"
export COINBASE_PRO_MAKER_FEE="your_coinbase_pro_api_maker_fee"
export COINBASE_PRO_TAKER_FEE="your_coinbase_pro_api_taker_fee"
export PERIOD_ALPHA="2021-06-02T08:00:00+00:00"
export PERIOD_OMEGA="2021-06-03T22:00:00+00:00"
export PERIOD_DURATION="24h00m00s"
```

### yml 
When you're ready to tune simulations and tweak patterns, create a `config.yml` configuration file like the following.
```yaml
# Coinbase Pro, with maker & taker fees
cbp:
  key:
  pass:
  secret:
  fees:
    maker:
    taker:

# Product selection and pattern criteria
patterns:
  - id: SKL-USD
    delta: .01
  - id: NU-USD
    gain: .0273
  - id: OMG-USD
    loss: .0473
  - id: TRB-USD
    size: 1.25

period:
    # A time frame for running the command 
    alpha: 2021-06-02T00:00:00+00:00
    omega: 2021-06-03T23:59:59+00:00
    # The amount of time to run the command
    duration: 24h0m0s
```
You'll need to place this file in the nuchal project directory or define config file location through the cli.

## Commands

### report
```shell
# Prints USD, Cryptocurrency, and total value of the configured Coinbase Pro account.
# Also prints position and trading information, namely size, value, balance and holds.
nuchal report -c /Users/${USER}/config.yml
```
![report example][10]

### sim
```shell
# Prints a simulation result report and serves a local website to host graphical report analysis.
nuchal sim

# Prints a simulation result report where the net gain for each product simulation was greater than zero.
nuchal sim -t --no-losers

# Prints a simulation result report where the net gain for each product simulation was greater than zero and also 
# where the amount of positions trading are zero.	
nuchal sim -w --winners-only
```
![sim example][12]
![chart example][14]

### trade
```shell
# Trade buys & sells products at prices or at times that meet or exceed pattern criteria, for a specified duration.
nuchal trade  --usd XLM,TRB,SKL,STORJ

# Hold creates a limit entry order at the goal price for every active trading position in your available balance.
nuchal trade --hold

# Sell all available positions (active trades) at prices or at times that meet or exceed pattern criteria.
nuchal trade --sell

# Sell all available positions (active trades) at the current market price. Will not sell holds.
nuchal trade --exit
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
docker compose -f build/docker-compose.yml down
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