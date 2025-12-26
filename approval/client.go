package approval

import (
	"encoding/json"
	"log"
	"time"

	apisdk "github.com/OpenBlockResource/openblock-api-sdk-go"
)

type Client struct {
	ApiKey        string
	ApiSecret     string
	apiClient     *apisdk.Client
	WalletInfoMap map[string]*WalletInfo
}

type WalletInfo struct {
	WalletId         string
	WalletName       string
	WalletAddressMap map[string]string
	IsHDWallet       bool
	HDWalletList     []*WalletInfo
}

func NewClient(apiKey, apiSecret string) *Client {
	client := apisdk.NewClient(apiKey, apiSecret, 10*time.Second)
	return &Client{
		ApiKey:        apiKey,
		ApiSecret:     apiSecret,
		apiClient:     client,
		WalletInfoMap: make(map[string]*WalletInfo),
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

func (c *Client) NewApproval(hdWalletId, action string, txInfo *apisdk.TXInfo, note string, expiredSeconds int32) (*apisdk.RespNewApproval, error) {
	txInfoJson, _ := json.Marshal(txInfo)
	log.Printf("NewApproval, hdWalletId: %s, action: %s, txInfo: %s, expiredSec: %d", hdWalletId, action, string(txInfoJson), expiredSeconds)
	return c.apiClient.CompanyWallet.NewApproval(&apisdk.ParamNewApproval{
		HDWalletID:     hdWalletId,
		Action:         action,
		TXInfo:         *txInfo,
		Note:           note,
		ExpiredTimeout: expiredSeconds,
	})
}

func (c *Client) GetWalletInfo() (*[]*WalletInfo, error) {
	var walletInfos []*WalletInfo
	mainWalletInfo, err := c.GetHDWalletInfo("-")
	if err != nil {
		return nil, err
	}
	walletInfos = append(walletInfos, mainWalletInfo)

	for _, hdWallet := range mainWalletInfo.HDWalletList {
		walletInfo, err := c.GetHDWalletInfo(hdWallet.WalletId)
		if err != nil {
			return nil, err
		}

		walletInfo.WalletName = hdWallet.WalletName
		walletInfos = append(walletInfos, walletInfo)
		time.Sleep(1 * time.Second)
	}
	return &walletInfos, nil
}

func (c *Client) GetHDWalletInfo(hdWalletId string) (*WalletInfo, error) {
	if walletInfo, ok := c.WalletInfoMap[hdWalletId]; ok && walletInfo != nil {
		return walletInfo, nil
	}
	log.Printf("get wallet info: %s", hdWalletId)

	walletInfo := WalletInfo{
		IsHDWallet:       true,
		WalletAddressMap: map[string]string{},
		WalletId:         hdWalletId,
	}
	if hdWalletId == "-" {
		walletInfo.IsHDWallet = false
		resp, err := c.apiClient.CompanyWallet.GetCompanyWalletInfo()
		if err != nil {
			return nil, err
		}
		walletInfo.WalletId = resp.Data.CompanyWalletInfo.CompanyWalletID
		walletInfo.WalletName = resp.Data.CompanyWalletInfo.WalletName

		for _, hdWallet := range resp.Data.HDWalletList {
			walletInfo.HDWalletList = append(walletInfo.HDWalletList, &WalletInfo{
				WalletId:   hdWallet.HDWalletID,
				WalletName: hdWallet.WalletName,
			})
		}

		for _, address := range resp.Data.AddressList {
			walletInfo.WalletAddressMap[address.Chain] = address.Address
		}

	} else {
		resp, err := c.apiClient.CompanyWallet.GetCompanyWalletHDWalletAddress(&apisdk.ParamGetCompanyWalletHDWalletAddress{
			HDWalletID: hdWalletId,
		})
		if err != nil {
			return nil, err
		}
		for _, address := range resp.Data.AddressList {
			walletInfo.WalletAddressMap[address.Chain] = address.Address
		}
	}

	c.WalletInfoMap[hdWalletId] = &walletInfo
	return &walletInfo, nil
}
