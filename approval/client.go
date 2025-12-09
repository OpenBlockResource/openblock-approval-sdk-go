package approval

import (
	"time"

	apisdk "github.com/OpenBlockResource/openblock-api-sdk-go"
)

type Client struct {
	ApiKey    string
	ApiSecret string
	apiClient *apisdk.Client
}

func NewClient(apiKey, apiSecret string) *Client {
	client := apisdk.NewClient(apiKey, apiSecret, 10*time.Second)
	return &Client{
		ApiKey:    apiKey,
		ApiSecret: apiSecret,
		apiClient: client,
	}
}

func (c *Client) GetApprovals(status string) (*apisdk.RespApprovals, error) {
	return c.apiClient.CompanyWallet.GetApprovals(&apisdk.ParamGetApprovals{
		Page:   1,
		Limit:  20,
		Status: status,
	})
}

func (c *Client) GetSponsoredApprovals(recordId string) (*apisdk.RespApprovalsV2, error) {
	return c.apiClient.CompanyWallet.GetApprovalsV2(&apisdk.ParamGetApprovalsV2{
		Page:     1,
		Limit:    20,
		ListType: "sponsor",
		RecordID: recordId,
	})
}

func (c *Client) AggreeApproval(approvalId string, agree bool) (*apisdk.RespAgreeApproval, error) {
	agreeStr := "reject"
	if agree {
		agreeStr = "agree"
	}
	return c.apiClient.CompanyWallet.AgreeApproval(&apisdk.ParamAgreeApproval{
		RecordID: approvalId,
		Agree:    agreeStr,
	})
}

func (c *Client) NewApproval(action string, txInfo *apisdk.TXInfo, note string) (*apisdk.RespNewApproval, error) {
	return c.apiClient.CompanyWallet.NewApproval(&apisdk.ParamNewApproval{
		Action: action,
		TXInfo: *txInfo,
		Note:   note,
	})
}
