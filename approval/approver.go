package approval

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

type ApprovalParams struct {
	MatchParams  []VerifyParams
	VerifyParams []VerifyParams
}

type VerifyParams struct {
	Path  string
	Value string
	Rule  string
}

type ApproveResults struct {
	ApprovalId string
	Approved   bool
	Action     string
	TxInfo     string
	HdWalletID string
}

func AutoAprove(client *Client, approvalParams *[]ApprovalParams) ([]ApproveResults, error) {
	apprs, err := client.GetApprovals("ING")
	if err != nil {
		return nil, err
	}
	var approveResult []ApproveResults
	for _, appr := range apprs.Data {
		if appr.Status != "ING" {
			continue
		}

		txInfoMap := convertTxInfoToMap(appr.ExtraData.Txinfo)
		var approveParams *ApprovalParams
		for _, params := range *approvalParams {
			approveParams = &params
			for _, param := range params.MatchParams {
				if !CheckParam(txInfoMap, param) {
					approveParams = nil
					break
				}
			}
			if approveParams != nil { //matched, break
				break
			}
		}
		if approveParams == nil {
			continue
		}

		agree := true
		for _, param := range approveParams.VerifyParams {
			if !CheckParam(convertTxInfoToMap(appr.ExtraData.Txinfo), param) {
				agree = false
				break
			}
		}

		res, err := client.AggreeApproval(appr.RecordId, agree)
		if err != nil {
			return nil, err
		}
		txInfo, _ := json.Marshal(appr.ExtraData.Txinfo)
		approveResult = append(approveResult, ApproveResults{
			ApprovalId: res.Data.RecordId,
			Approved:   agree,
			Action:     appr.ActionType,
			TxInfo:     string(txInfo),
			HdWalletID: appr.HDWalletID,
		})

	}
	return approveResult, nil
}

// 根据不同的Type和Rule检查参数
func CheckParam(txInfo map[string]interface{}, param VerifyParams) bool {
	actualValue := getValueByPath(txInfo, param.Path)
	if actualValue == nil {
		return false
	}

	switch actual := actualValue.(type) {
	case string:
		actualDecimal, err := decimal.NewFromString(actual)
		if err != nil {
			return checkByRule(actual, param.Value, param.Rule)
		} else {
			return checkDecimalRule(actualDecimal, param.Value, param.Rule)
		}

	case []interface{}:
		// 检查金额
		return checkListRule(actual, param.Value, param.Rule)
	default:
		// 默认精确匹配
		return false
	}
}

func checkListRule(actual []interface{}, expectedValue, rule string) bool {
	switch rule {
	case "length":
		// 校验列表长度
		expectedLen, err := strconv.Atoi(expectedValue)
		if err != nil {
			return false
		}
		return len(actual) == expectedLen
	case "minLength":
		// 校验列表最小长度
		minLen, err := strconv.Atoi(expectedValue)
		if err != nil {
			return false
		}
		return len(actual) >= minLen
	case "maxLength":
		// 校验列表最大长度
		maxLen, err := strconv.Atoi(expectedValue)
		if err != nil {
			return false
		}
		return len(actual) <= maxLen
	case "contains":
		// 检查列表是否包含指定值
		for _, item := range actual {
			if v, ok := item.(string); ok && v == expectedValue {
				return true
			}
		}
		return false
	case "notContains":
		// 检查列表是否不包含指定值
		for _, item := range actual {
			if v, ok := item.(string); ok && v == expectedValue {
				return false
			}
		}
		return true
	default:
		// 未知规则，默认返回false
		return false
	}
}

// 根据Rule规则检查值
func checkByRule(actualValue, expectedValue, rule string) bool {
	switch rule {
	case "exact":
		// 精确匹配
		return actualValue == expectedValue
	case "contains":
		// 包含匹配
		return strings.Contains(actualValue, expectedValue)
	case "prefix":
		// 前缀匹配
		return strings.HasPrefix(actualValue, expectedValue)
	case "suffix":
		// 后缀匹配
		return strings.HasSuffix(actualValue, expectedValue)
	case "regex":
		// 正则表达式匹配
		matched, err := regexp.MatchString(expectedValue, actualValue)
		if err != nil {
			return false
		}
		return matched
	default:
		// 默认为精确匹配
		return actualValue == expectedValue
	}
}

// decimal数值比较规则
func checkDecimalRule(actualValue decimal.Decimal, expectedValue, rule string) bool {
	expected, err := decimal.NewFromString(expectedValue)
	if err != nil && rule != "range" {
		return false
	}

	switch rule {
	case "eq":
		return actualValue.Equal(expected)
	case "gt":
		return actualValue.GreaterThan(expected)
	case "gte":
		return actualValue.GreaterThanOrEqual(expected)
	case "lt":
		return actualValue.LessThan(expected)
	case "lte":
		return actualValue.LessThanOrEqual(expected)
	case "range":
		// expectedValue 应该是 "min,max" 格式
		parts := strings.Split(expectedValue, ",")
		if len(parts) != 2 {
			return false
		}
		min, err1 := decimal.NewFromString(parts[0])
		max, err2 := decimal.NewFromString(parts[1])
		if err1 != nil || err2 != nil {
			return false
		}
		return actualValue.GreaterThanOrEqual(min) && actualValue.LessThanOrEqual(max)
	default:
		return actualValue.Equal(expected)
	}
}

// 将txInfo转换为map[string]interface{}
func convertTxInfoToMap(txInfo interface{}) map[string]interface{} {
	// 如果已经是map类型，直接返回
	if m, ok := txInfo.(map[string]interface{}); ok {
		return m
	}

	// 尝试通过json序列化和反序列化转换为map
	jsonBytes, err := json.Marshal(txInfo)
	if err != nil {
		return nil
	}

	var resultMap map[string]interface{}
	err = json.Unmarshal(jsonBytes, &resultMap)
	if err != nil {
		return nil
	}

	return resultMap
}

// 通过路径从map中获取值，支持嵌套路径，如 "ExtraData.Txinfo.chain"
// 也支持数组索引，如 "transactions.0.to"
func getValueByPath(m map[string]interface{}, path string) interface{} {
	// 按点号分割路径
	keys := strings.Split(path, ".")

	var currentValue interface{} = m

	// 逐级深入查找
	for _, key := range keys {
		switch current := currentValue.(type) {
		case map[string]interface{}:
			// 如果当前值是map类型
			var ok bool
			currentValue, ok = current[key]
			if !ok {
				// 键不存在
				return nil
			}
		case []interface{}:
			// 如果当前值是数组类型
			index, ok := parseIndex(key)
			if !ok {
				// 不是有效的索引
				return nil
			}
			if index >= len(current) {
				// 索引越界
				return nil
			}
			if index < 0 {
				// 索引小于0
				index = len(current) + index
			}
			currentValue = current[index]
		default:
			// 当前值不是map或数组类型，无法继续深入
			return nil
		}
	}

	return currentValue
}

func parseIndex(key string) (int, bool) {
	index, err := strconv.Atoi(key)
	if err != nil {
		return 0, false
	}
	return index, true
}
