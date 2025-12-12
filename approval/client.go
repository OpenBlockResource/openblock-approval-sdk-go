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

type WalletInfo struct {
	WalletId         string
	WalletName       string
	WalletAddressMap map[string]string
	IsHDWallet       bool
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

func (c *Client) NewApproval(hdWalletId, action string, txInfo *apisdk.TXInfo, note string) (*apisdk.RespNewApproval, error) {
	return c.apiClient.CompanyWallet.NewApproval(&apisdk.ParamNewApproval{
		HDWalletID: hdWalletId,
		Action:     action,
		TXInfo:     *txInfo,
		Note:       note,
	})
}

func (c *Client) GetWalletInfo() (*[]WalletInfo, error) {
	resp, err := c.apiClient.CompanyWallet.GetCompanyWalletInfo()
	if err != nil {
		return nil, err
	}
	var walletInfos []WalletInfo
	walletInfos = append(walletInfos, WalletInfo{
		IsHDWallet:       false,
		WalletAddressMap: map[string]string{},
		WalletId:         resp.Data.CompanyWalletInfo.CompanyWalletID,
		WalletName:       resp.Data.CompanyWalletInfo.WalletName,
	})
	for _, address := range resp.Data.AddressList {
		walletInfos[0].WalletAddressMap[address.Chain] = address.Address
	}
	for _, hdWallet := range resp.Data.HDWalletList {
		walletInfo := WalletInfo{
			IsHDWallet:       true,
			WalletAddressMap: map[string]string{},
			WalletId:         hdWallet.HDWalletID,
			WalletName:       hdWallet.WalletName,
		}
		resp2, err := c.apiClient.CompanyWallet.GetCompanyWalletHDWalletAddress(&apisdk.ParamGetCompanyWalletHDWalletAddress{
			HDWalletID: hdWallet.HDWalletID,
		})
		if err != nil {
			return nil, err
		}
		for _, address := range resp2.Data.AddressList {
			walletInfo.WalletAddressMap[address.Chain] = address.Address
		}
		walletInfos = append(walletInfos, walletInfo)
		time.Sleep(1 * time.Second)
	}
	return &walletInfos, nil
}
