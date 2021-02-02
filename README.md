# evm lab

light weight evm laboratory to test evm syntax, the usage is as simple as:

1. set initial state in config.json
2. create contract or call contract

## steps

```
# start server, set initial state as needed
$ go run main.go server start --cfg config.json

# create contract
$ go run main.go client create_contract --contract_path /Users/xuzhiqiang/Desktop/workspace/opensource/go_projects/eth-contracts/contracts/core/lock_proxy/LockProxy.sol --sender 71562b71999873db5b286df957af199ec94617f7
output {"Addr":"0x3a220f351252089d385b29beca14e27f204c2960","ErrMsg":""}

# call contract
$ go run main.go client call_contract --contract_path /Users/xuzhiqiang/Desktop/workspace/opensource/go_projects/eth-contracts/contracts/core/lock_proxy/LockProxy.sol --sender 71562b71999873db5b286df957af199ec94617f7 --receiver 0x3a220f351252089d385b29beca14e27f204c2960  --method setManagerProxy 0x05fF834dD5a7EDB437B061CB00108200bf4873D6
output {"Result":null,"ErrMsg":""}

# if something wrong(last character of sender is wrong)
$ go run main.go client call_contract --contract_path /Users/xuzhiqiang/Desktop/workspace/opensource/go_projects/eth-contracts/contracts/core/lock_proxy/LockProxy.sol --sender 71562b71999873db5b286df957af199ec94617f1 --receiver 0x3a220f351252089d385b29beca14e27f204c296a  --method setManagerProxy 0x05fF834dD5a7EDB437B061CB00108200bf4873D6
output {"Result":"CMN5oAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACBPd25hYmxlOiBjYWxsZXIgaXMgbm90IHRoZSBvd25lcg==","ErrMsg":"execution reverted"}

```