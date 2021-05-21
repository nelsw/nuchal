# nuchal
A project for evaluating & executing cryptocurrency trades on Coinbase Pro. 

> **nuchal**, *new⋅cull, nu⋅chal, /ˈn(y)o͞ok(ə)l/,*
> 
> "In marine biology, an exaggerated craniofacial trait responsible for defining rank and order among a species."

This project does the same - it recognizes market characteristics and trades the best (profitable) selection.

# Overview
The **goals** of this project are to:
- recognize trade opportunities based on statistical pattern analysis
- automate market and limit trading based on pattern recognition
- simulate automation results using historical ticker data
- provide a streaming summary of portfolio positions

## Build

## Configure
Before you get started, you'll need a [Coinbase Pro][1] API key and working installation of [Go][2] and [git][3]    .

### Single User
```shell
export COINBASE_PRO_KEY="your_coinbase_pro_api_key"
export COINBASE_PRO_PASSPHRASE="your_coinbase_pro_api_passphrase"
export COINBASE_PRO_SECRET="your_coinbase_pro_api_secret"
```

### Multiple Users

#### .sql

#### .json
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