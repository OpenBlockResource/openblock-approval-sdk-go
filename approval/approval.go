package approval

import (
	"encoding/json"
	"os"

	apisdk "github.com/OpenBlockResource/openblock-api-sdk-go"
)

type ApprovalWallet struct {
	Role           string
	ApiKey         string
	ApiSecret      string
	ApprovalParams []ApprovalParams
	TxInfo         *apisdk.TXInfo
	Client         *Client
}

/*
  - 初始化审批钱包
    @apiKey: openblock api key
    @apiSecret: openblock api secret
*/
func NewApprovalWallet(apiKey, apiSecret string) *ApprovalWallet {
	return &ApprovalWallet{
		ApiKey:         apiKey,
		ApiSecret:      apiSecret,
		ApprovalParams: []ApprovalParams{},
		Client:         NewClient(apiKey, apiSecret),
	}
}

/*
  - 发送审批交易
    @chainName: Solana/ETH/Benfen
    @txData: SOL base64交易
    返回值: txHash
*/
func (w *ApprovalWallet) SendApprovalTransaction(hdWalletId, chainName, txData string) (string, error) {
	return SendApprovalTransaction(w.Client, hdWalletId, chainName, txData)
}

/*
  - 只签名不发送审批交易
    @chainName: Solana/ETH/Benfen
    @txData: SOL base64交易
    返回值: txHash
*/
func (w *ApprovalWallet) SignApprovalTransaction(hdWalletId, chainName, txData string) (string, error) {
	return SignApprovalTransaction(w.Client, hdWalletId, chainName, txData)
}

/*
  - 签名审批交易/消息签名
    @chainName: Solana/ETH/Benfen
    @txInfo: 参考ob api接口
    返回值: txHash/签名
*/
func (w *ApprovalWallet) SendApprovalTxInfo(hdWalletId string, txInfo *apisdk.TXInfo) (string, error) {
	return SendApprovalTxInfo(w.Client, hdWalletId, txInfo)
}

/*
  - 发送审批消息签名
    @chainName: Solana/ETH/Benfen
    @message: SOL hex格式消息
    返回值: 签名
*/
func (w *ApprovalWallet) SignApprovalMessage(hdWalletId, chainName, message string) (string, error) {
	return SignApprovalMessage(w.Client, hdWalletId, chainName, message)
}

/*
- 自动审批交易
*/
func (w *ApprovalWallet) AutoApprove() error {
	_, err := AutoApprove(w.Client, &w.ApprovalParams)
	return err
}

/*
  - 签名交易
    返回值: txHash
*/
func (w *ApprovalWallet) AutoSign() error {
	return AutoSign(w.Client, &w.ApprovalParams)
}

func NewApprovalWalletFromJson(filePath string) (*ApprovalWallet, error) {
	// 读取JSON文件
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	w := ApprovalWallet{}
	if err := json.Unmarshal(data, &w); err != nil {
		return nil, err
	}
	w.Client = NewClient(w.ApiKey, w.ApiSecret)

	return &w, nil
}
