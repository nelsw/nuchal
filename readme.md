# nuchal
A program for evaluating & executing cryptocurrency trades on Coinbase Pro. 

# Overview
The **goals** of this program are to:
- recognize trade opportunities based on statistical pattern analysis
- automate market and limit trading based on pattern recognition
- simulate automation results using historical ticker data
- provide a streaming summary of portfolio positions

## Configuration
Before you get started, you'll need a Coinbase Pro API key and working installation of golang.

### Single User

### Multiple Users
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