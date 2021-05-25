# nuchal
**nuchal**, *new⋅cull, nu⋅chal, /ˈn(y)o͞ok(ə)l/,* - a program for trading cryptocurrency on Coinbase Pro for *1-n* users.

# Overview
The **goals** of this project are to:
- recognize trade opportunities based on statistical pattern analysis
- automate market and limit trading based on pattern recognition
- simulate automation results using historical ticker data
- provide a streaming summary of portfolio positions


## Setup
This project requires a [Coinbase Pro][1] API key and working installations of [Go][2], [git][3], and [Docker][4].

### Configure
Required environment variables.
```shell
export GOOS=linux 
export GOARCH=amd64
export PATH=${PATH}:/Users/${USER}/go/bin
```

#### Single User
```shell
export COINBASE_PRO_KEY="your_coinbase_pro_api_key"
export COINBASE_PRO_PASSPHRASE="your_coinbase_pro_api_passphrase"
export COINBASE_PRO_SECRET="your_coinbase_pro_api_secret"
```

#### Multiple Users

##### .json
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

### Build

```shell
git clone https://github.com/nelsw/nuchal.git &&
cd nuchal &&
go mod download &&
go mod tidy &&
go build -o build/nuchal main.go &&
go install
```

## Use

### Simulate
```shell
docker compose -p nuchal -f build/docker-compose.yml up -d

docker compose -f build/docker-compose.yml down
```
```shell
nuchal sim --serve

nuchal sim --coin 'ADA,MATIC,XTZ'

nuchal sim --user 'Carl Brutananadilewski'
```
### Trade

### Report

# License
This code is Copyright Connor Ross Van Elswyk and licensed under Apache License 2.0

[1]: https://pro.coinbase.com
[2]: https://golang.org/
[3]: https://git-scm.com/
[4]: https://git-scm.com/