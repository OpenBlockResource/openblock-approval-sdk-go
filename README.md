# OpenBlock Approval SDK

OpenBlock Approval SDK 是一个用于处理企业钱包交易自动审批的 Go 语言库。


## 编译运行、调用
- 编译：go build cmd/runner.go
- 脚本命令：
```bash
#Usage of ./runner:
#  -check-wallet string
#    	Check wallet information, e.g. -check-wallet=Solana,ETH
#   -config string
#     	Path to the configuration file (default "config.json")
#   -hd-wallet-id string
#     	ID of the HD wallet


./runner -config cmd/manager.json #管理员持续审批
./runner -config cmd/sol-transaction.json #发起sol交易
./runner -config cmd/eth-transaction.json -hd-wallet-id 0ced4ad982e84efdb282bd16b913459a #发起子钱包ETH交易
./runner -config cmd/evm-message.json #发起签名
./runner -check-wallet ETH,Solana #查看钱包ID和地址
```
- sdk调用:
```go
import "openblock-approval-sdk/approval"
...

//初始化
wallet, err := approval.NewApprovalWallet(apiKey, apiSecret)
...

//消息签名
hdWalletId := "6861e10455f84ab0bb130425c3440c60" //hd钱包id，如果是主钱包，则为空
res, err := wallet.SignApprovalMessage(hdWalletId, "Solana", "abcdef1234567890")
...

//发送交易
res, err := wallet.SendApprovalTransaction(hdWalletId, "Solana", txData)

//只签名不发送交易
res, err := wallet.SignApprovalTransaction(hdWalletId, "Solana", txData)
...

//evm 交易
txInfoJson := `{
    "from":                 "from address",
    "to":                   "to address",
    "value":                "0.01",
    "data":                 "0x...", //hex data
    "chain":                "ETH",
    "gasLimit":             "21000",
    "gasPrice":             "1", // 1 gwei
    "nonce":                "0",
    "eip1559": 				true, //true时手续费使用下面两个字段
    "maxFeePerGas":         "5", //5 gwei
    "maxPriorityFeePerGas": "1", // 1 gwei
}`
//chain: ETH, BSC, Polygon, Arbitrum, Optimism, Avalanche, Fantom
wallet.SendApprovalTransaction(hdWalletId, "ETH", txInfoJson)

//... 其他参考cmd下的交易模板

```


## Role角色
- initiator：发起人
  - 发起审批，可以在config.json中配置txInfo，根据配置发起交易/消息签名，具体字段参考：https://docs.openblock.com/zh-Hans/OpenBlock/API/Enterprise%20Wallet/#%E5%88%9B%E5%BB%BA%E4%BA%A4%E6%98%93%E7%9B%B8%E5%85%B3%E5%AE%A1%E6%89%B9
- approver：审批人
  - 自动查询审批列表，将MatchParams匹配到的审批，按照VerifyParams进行审批
- manager：管理员
  - 自动查询审批列表，将MatchParams匹配到的审批，按照VerifyParams进行审批，并调用docker完成mpc签名
  - docker部署：
    - https://docs.openblock.com/zh-Hans/OpenBlock/API/Enterprise%20Wallet/#docker-api
- 可以和openblock端上交叉使用，如：web端人工发起，脚本自动审批，或者脚本发起，web端人工审批



## MatchParams/VerifyParams 规则说明

先根据MatchParams匹配txInfo（交易或者消息签名），对匹配到的txInfo根据VerifyParams进行审批。
MatchParams 和VerifyParams相同结构。
VerifyParams 结构体用于定义交易验证参数，包含以下字段：

- `Path`: 对应txInfo中字段的路径
- `Value`: 规则要用的的验证值
- `Rule`: 验证规则

### Path 规则

1. map结构通过.分割的字段进行json路径导航，如：`transfer.amount`
2. list结构支持数字索引，如：transfer.0.amount


### 支持的 Rule 规则

#### 字符串类型规则

- `exact` - 精确匹配
- `contains` - 包含匹配
- `prefix` - 前缀匹配
- `suffix` - 后缀匹配
- `regex` - 正则表达式匹配

#### 数值类型规则

- `eq` - 等于
- `gt` - 大于
- `gte` - 大于等于
- `lt` - 小于
- `lte` - 小于等于
- `range` - 范围匹配（value格式："min,max"）

#### 列表类型规则

- `length` - 长度等于
- `minLength` - 最小长度
- `maxLength` - 最大长度
- `contains` - 包含
- `notContains` - 不包含

