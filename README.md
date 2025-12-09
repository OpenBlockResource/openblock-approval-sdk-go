# OpenBlock Approval SDK

OpenBlock Approval SDK 是一个用于处理区块链交易审批的 Go 语言库。

## 编译运行、调用
- 编译：go build cmd/runner.go
- 脚本运行：./runner -config=config.yaml
- sdk调用:
```go
import "openblock-approval-sdk/approval"
...

wallet, err := approval.NewApprovalWallet(apiKey, apiSecret)
...

res, err := wallet.SendApprovalTransaction("Solana", txData)
...
```

## Role角色
- initiator：发起人
  - 发起审批，可以在config.yaml中配置txInfo，具体字段参考：https://docs.openblock.com/zh-Hans/OpenBlock/API/Enterprise%20Wallet/#%E5%88%9B%E5%BB%BA%E4%BA%A4%E6%98%93%E7%9B%B8%E5%85%B3%E5%AE%A1%E6%89%B9
- approver：审批人
  - 自动查询审批列表，将MatchParams匹配到的审批，按照VerifyParams进行审批
- manager：管理员
  - 自动查询审批列表，将MatchParams匹配到的审批，按照VerifyParams进行审批，并调用docker完成mpc签名
  - docker部署： https://docs.openblock.com/zh-Hans/OpenBlock/API/Enterprise%20Wallet/#docker-api

## MatchParams/VerifyParams 规则说明

MatchParams 用于交易匹配, 和VerifyParams相同结构。
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
- `range` - 范围匹配（格式："min,max"）

#### 列表类型规则

- `length` - 长度等于
- `minLength` - 最小长度
- `maxLength` - 最大长度
- `contains` - 包含
- `notContains` - 不包含

