package approval

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestCheckParam(t *testing.T) {
	// 构造测试用的txInfo数据
	txInfo := map[string]interface{}{
		"chain":     "ethereum",
		"from":      "0x123abc",
		"to":        "0x456def",
		"amount":    "100.50",
		"fee":       "1.25",
		"contracts": []interface{}{"contract1", "contract2"},
		"nested": map[string]interface{}{
			"value": "nested_value",
			"data": []interface{}{
				map[string]interface{}{"address": "0xabc"},
				map[string]interface{}{"address": "0xdef"},
			},
		},
	}

	t.Run("字符串精确匹配", func(t *testing.T) {
		param := VerifyParams{
			Path:  "chain",
			Value: "ethereum",
			Rule:  "exact",
		}
		result := CheckParam(txInfo, param)
		assert.True(t, result)
	})

	t.Run("字符串包含匹配", func(t *testing.T) {
		param := VerifyParams{
			Path:  "from",
			Value: "123",
			Rule:  "contains",
		}
		result := CheckParam(txInfo, param)
		assert.True(t, result)
	})

	t.Run("字符串前缀匹配", func(t *testing.T) {
		param := VerifyParams{
			Path:  "to",
			Value: "0x456",
			Rule:  "prefix",
		}
		result := CheckParam(txInfo, param)
		assert.True(t, result)
	})

	t.Run("字符串后缀匹配", func(t *testing.T) {
		param := VerifyParams{
			Path:  "from",
			Value: "abc",
			Rule:  "suffix",
		}
		result := CheckParam(txInfo, param)
		assert.True(t, result)
	})

	t.Run("正则表达式匹配", func(t *testing.T) {
		param := VerifyParams{
			Path:  "chain",
			Value: "^eth.*um$",
			Rule:  "regex",
		}
		result := CheckParam(txInfo, param)
		assert.True(t, result)
	})

	t.Run("decimal数值相等匹配", func(t *testing.T) {
		param := VerifyParams{
			Path:  "amount",
			Value: "100.50",
			Rule:  "eq",
		}
		result := CheckParam(txInfo, param)
		assert.True(t, result)
	})

	t.Run("decimal数值大于匹配", func(t *testing.T) {
		param := VerifyParams{
			Path:  "amount",
			Value: "100",
			Rule:  "gt",
		}
		result := CheckParam(txInfo, param)
		assert.True(t, result)
	})

	t.Run("decimal数值小于匹配", func(t *testing.T) {
		param := VerifyParams{
			Path:  "amount",
			Value: "101",
			Rule:  "lt",
		}
		result := CheckParam(txInfo, param)
		assert.True(t, result)
	})

	t.Run("decimal数值范围匹配", func(t *testing.T) {
		param := VerifyParams{
			Path:  "fee",
			Value: "1,2",
			Rule:  "range",
		}
		result := CheckParam(txInfo, param)
		assert.True(t, result)
	})

	t.Run("列表长度匹配", func(t *testing.T) {
		param := VerifyParams{
			Path:  "contracts",
			Value: "2",
			Rule:  "length",
		}
		result := CheckParam(txInfo, param)
		assert.True(t, result)
	})

	t.Run("列表最小长度匹配", func(t *testing.T) {
		param := VerifyParams{
			Path:  "contracts",
			Value: "1",
			Rule:  "minLength",
		}
		result := CheckParam(txInfo, param)
		assert.True(t, result)
	})

	t.Run("列表最大长度匹配", func(t *testing.T) {
		param := VerifyParams{
			Path:  "contracts",
			Value: "3",
			Rule:  "maxLength",
		}
		result := CheckParam(txInfo, param)
		assert.True(t, result)
	})

	t.Run("列表包含匹配", func(t *testing.T) {
		param := VerifyParams{
			Path:  "contracts",
			Value: "contract1",
			Rule:  "contains",
		}
		result := CheckParam(txInfo, param)
		assert.True(t, result)
	})

	t.Run("列表不包含匹配", func(t *testing.T) {
		param := VerifyParams{
			Path:  "contracts",
			Value: "contract3",
			Rule:  "notContains",
		}
		result := CheckParam(txInfo, param)
		assert.True(t, result)
	})

	t.Run("嵌套路径值匹配", func(t *testing.T) {
		param := VerifyParams{
			Path:  "nested.value",
			Value: "nested_value",
			Rule:  "exact",
		}
		result := CheckParam(txInfo, param)
		assert.True(t, result)
	})

	t.Run("数组索引路径匹配", func(t *testing.T) {
		param := VerifyParams{
			Path:  "nested.data.0.address",
			Value: "0xabc",
			Rule:  "exact",
		}
		result := CheckParam(txInfo, param)
		assert.True(t, result)
	})

	t.Run("路径不存在的情况", func(t *testing.T) {
		param := VerifyParams{
			Path:  "nonexistent.path",
			Value: "any_value",
			Rule:  "exact",
		}
		result := CheckParam(txInfo, param)
		assert.False(t, result)
	})

	t.Run("不匹配的情况", func(t *testing.T) {
		param := VerifyParams{
			Path:  "chain",
			Value: "bitcoin",
			Rule:  "exact",
		}
		result := CheckParam(txInfo, param)
		assert.False(t, result)
	})
}

func TestCheckByRule(t *testing.T) {
	t.Run("默认匹配规则", func(t *testing.T) {
		result := checkByRule("test", "test", "unknown_rule")
		assert.True(t, result)

		result = checkByRule("test", "different", "unknown_rule")
		assert.False(t, result)
	})
}

func TestCheckDecimalRule(t *testing.T) {
	value := decimal.NewFromFloat(100.50)

	t.Run("无效范围格式", func(t *testing.T) {
		result := checkDecimalRule(value, "invalid_range", "range")
		assert.False(t, result)
	})

	t.Run("范围值不足", func(t *testing.T) {
		result := checkDecimalRule(value, "100", "range")
		assert.False(t, result)
	})

	t.Run("范围值过多", func(t *testing.T) {
		result := checkDecimalRule(value, "100,200,300", "range")
		assert.False(t, result)
	})

	t.Run("无效最小值", func(t *testing.T) {
		result := checkDecimalRule(value, "invalid,200", "range")
		assert.False(t, result)
	})

	t.Run("无效最大值", func(t *testing.T) {
		result := checkDecimalRule(value, "100,invalid", "range")
		assert.False(t, result)
	})
}

func TestCheckListRule(t *testing.T) {
	list := []interface{}{"item1", "item2", "item3"}

	t.Run("无效长度值", func(t *testing.T) {
		result := checkListRule(list, "invalid", "length")
		assert.False(t, result)
	})

	t.Run("无效最小长度值", func(t *testing.T) {
		result := checkListRule(list, "invalid", "minLength")
		assert.False(t, result)
	})

	t.Run("无效最大长度值", func(t *testing.T) {
		result := checkListRule(list, "invalid", "maxLength")
		assert.False(t, result)
	})

	t.Run("未知规则", func(t *testing.T) {
		result := checkListRule(list, "any_value", "unknown_rule")
		assert.False(t, result)
	})
}
