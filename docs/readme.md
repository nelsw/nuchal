# nuchal
**nuchal**, *new⋅cull, nu⋅chal, /ˈn(y)o͞ok(ə)l/,* - a program for trading cryptocurrency on Coinbase Pro for *1-n* users.

# Overview
The **goals** of this project are to:
- recognize trade opportunities based on statistical pattern analysis
- automate market and limit trading based on pattern recognition
- simulate automation results using historical ticker data
- provide a streaming summary of portfolio positions


## Build
Before you get started, you'll need a [Coinbase Pro][1] API key and working installation of [Go][2] and [git][3].

```shell
git clone https://github.com/nelsw/nuchal.git && 
cd nuchal &&
GOOS=linux GOARCH=amd64 && 
go build -o build/nuchal main.go && 
PATH=${PATH}:/Users/${USER}/go/bin && 
go install
```

### Single User
```shell
export PORT=8080
export MODE="DEBUG"
export DURATION=24h
export USER=${USER}
export COINBASE_PRO_KEY="your_coinbase_pro_api_key"
export COINBASE_PRO_PASSPHRASE="your_coinbase_pro_api_passphrase"
export COINBASE_PRO_SECRET="your_coinbase_pro_api_secret"
```

### Multiple Users

#### .json
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

## Use

### Simulate

### Trade

### Report

# License
This code is Copyright Connor Ross Van Elswyk and licensed under Apache License 2.0

[1]: https://pro.coinbase.com
[2]: https://golang.org/
[3]: https://git-scm.com/