package approval

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	apisdk "github.com/OpenBlockResource/openblock-api-sdk-go"
	"github.com/fardream/go-bcs/bcs"
	solana "github.com/gagliardetto/solana-go"
	"github.com/minio/blake2b-simd"
)

const SOLANA = "Solana"
const ETHEREUM = "ETH"
const BSC = "BSC"
const POLYGON = "Polygon"
const ARBITRUM = "Arbitrum"
const OPTIMISM = "Optimism"
const AVALANCHE = "Avalanche"
const FANTOM = "Fantom"
const BENFEN = "Benfen"
const BENFEN_TESTNET = "BenfenTEST"

func SendApprovalTransaction(client *Client, hdWalletId, chainName, txData string) (string, error) {
	txInfo, err := BuildTxInfo(chainName, txData, true)
	if err != nil {
		return "", err
	}
	return SendApprovalTxInfo(client, hdWalletId, txInfo)
}

func SignApprovalTransaction(client *Client, hdWalletId, chainName, txData string) (string, error) {
	txInfo, err := BuildTxInfo(chainName, txData, false)
	if err != nil {
		return "", err
	}
	return SendApprovalTxInfo(client, hdWalletId, txInfo)
}

func BuildTxInfo(chainName, txData string, needSend bool) (*apisdk.TXInfo, error) {
	var txInfo *apisdk.TXInfo
	switch chainName {
	case SOLANA:
		transaction, err := solana.TransactionFromBase64(txData)
		if err != nil {
			return nil, err
		}
		txInput := map[string]any{
			"recent_blockhash": transaction.Message.RecentBlockhash.String(),
			"header": map[string]any{
				"numReadonlySignedAccounts":   transaction.Message.Header.NumReadonlySignedAccounts,
				"numReadonlyUnsignedAccounts": transaction.Message.Header.NumReadonlyUnsignedAccounts,
				"numRequiredSignatures":       transaction.Message.Header.NumRequiredSignatures,
			},
			"staticAccountKeys":    transaction.Message.AccountKeys.ToBase58(),
			"compiledInstructions": ConvertCompiledInstructions(transaction.Message.Instructions),
			"addressTableLookups":  transaction.Message.AddressTableLookups,
		}

		txInfo = &apisdk.TXInfo{
			RecentBlockHash: transaction.Message.RecentBlockhash.String(),
			TxPayload: []any{
				txInput,
			},
			TransactionType: "contract",
			ActiveTokenEnum: 1,
		}
		if needSend {
			txInfo.BridgeMethod = "solana_signTransaction"
		}

	case ETHEREUM, POLYGON, ARBITRUM, OPTIMISM, AVALANCHE, FANTOM, BSC:
		txInfo = &apisdk.TXInfo{}
		if err := json.Unmarshal([]byte(txData), &txInfo); err != nil {
			return nil, fmt.Errorf("evm txData format error: %s", err)
		}
		txInfo.TransactionType = "native"
		if needSend {
			txInfo.BridgeMethod = "eth_signTransaction"
		}

	case BENFEN, BENFEN_TESTNET:
		txInfo = &apisdk.TXInfo{
			Data:            txData,
			Payload:         map[string]string{},
			TransactionType: "native",
		}
		if needSend {
			txInfo.BridgeMethod = "bfc_signTransaction"
		}

	default:
		return nil, fmt.Errorf("not supported")
	}
	txInfo.Chain = chainName
	txInfo.Method = txInfo.BridgeMethod

	return txInfo, nil
}

func SignApprovalMessage(client *Client, hdWalletId, chainName, message string) (string, error) {
	var txInfo *apisdk.TXInfo
	hrMessage := message
	switch chainName {
	case SOLANA:
		m, err := hex.DecodeString(message)
		if err != nil {
			return "", fmt.Errorf("invalid hex message: %s", err)
		}
		hrMessage = string(m)

		txInfo = &apisdk.TXInfo{
			Chain:  chainName,
			Method: "solana_signMessage",
			Msg: &apisdk.Msg{
				SignMsg:     message,
				Message:     hrMessage,
				OriginalMsg: message,
			},
		}

	case ETHEREUM, POLYGON, ARBITRUM, OPTIMISM, AVALANCHE, FANTOM, BSC:
		method := "eth_signTypedData_v4"
		if message[:2] == "0x" || !(message[0] == '{' || message[0] == '[') {
			method = "personal_sign"
			if message[:2] == "0x" {
				m, err := hex.DecodeString(message)
				if err != nil {
					return "", fmt.Errorf("invalid hex message: %s", err)
				}
				hrMessage = string(m)
			}
		}

		txInfo = &apisdk.TXInfo{
			Chain:  chainName,
			Method: method,
			Msg: &apisdk.Msg{
				SignMsg:     message,
				Message:     hrMessage,
				OriginalMsg: message,
			},
		}

	case BENFEN, BENFEN_TESTNET:
		s1, err := hex.DecodeString(message)
		if err != nil {
			return "", fmt.Errorf("invalid hex message: %s", err)
		}
		hrMessage = string(s1)

		s2, err := bcs.Marshal(s1)
		if err != nil {
			return "", fmt.Errorf("invalid bsc message: %s", err)
		}
		s3 := append([]byte{3, 0, 0}, s2...)
		s4 := blake2b.Sum256(s3)

		txInfo = &apisdk.TXInfo{
			Chain:  chainName,
			Method: "bfc_signMessage",
			Msg: &apisdk.Msg{
				SignMsg:     hex.EncodeToString(s4[:]),
				Message:     hrMessage,
				OriginalMsg: message,
			},
		}

	default:
		return "", fmt.Errorf("not supported")
	}

	return SendApprovalTxInfo(client, hdWalletId, txInfo)
}

func SendApprovalTxInfo(client *Client, hdWalletId string, txInfo *apisdk.TXInfo) (string, error) {
	expiredSeconds := int32(0)
	action := "TRANSACTION"
	if strings.HasSuffix(txInfo.BridgeMethod, "_signTransaction") || //只签名不发送交易
		txInfo.TransactionType == "contract" {
		expiredSeconds = int32(300)
		action = "TRANSACTION_CONTRACT_INTERACTION"

	} else if txInfo.Msg != nil && txInfo.Msg.SignMsg != "" { //消息签名
		action = "TRANSACTION_SIGNATURE"
	}

	appr, err := client.NewApproval(hdWalletId, action, txInfo, "", expiredSeconds)
	if err != nil {
		return "", err
	}
	recordId := appr.Data.OriginRecordId

	for i := 0; i < 15; i++ {
		apprs, err := client.GetSponsoredApprovals(recordId)
		if err != nil {
			return "", fmt.Errorf("GetSponsoredApprovals error: %v", err)
		}
		for _, appr := range apprs.Data.Data {
			if appr.RecordID != recordId {
				continue
			}

			switch appr.Status {
			case "AGREE":
				res := appr.TxHash
				if strings.HasSuffix(txInfo.BridgeMethod, "_signTransaction") { //只签名不发送交易
					if appr.ExtraData.CustomData == "" {
						return "", fmt.Errorf("customData is empty, recordId: %s", recordId)
					}

					var resData any
					var rawTx string
					json.Unmarshal([]byte(appr.ExtraData.CustomData), &resData)
					if v, ok := resData.(map[string]any); ok && v["data"] != nil {
						if vv, ok := v["data"].(string); ok {
							rawTx = vv
						} else if vv, ok := v["data"].([]any); ok {
							rawTx = vv[0].(string)
						} else {
							return "", fmt.Errorf("invalid customData, recordId: %s", recordId)
						}
					}
					if rawTx == "" {
						return "", fmt.Errorf("rawTx is empty, recordId: %s", recordId)
					}

					switch txInfo.Chain {
					case BENFEN, BENFEN_TESTNET:
						var suiTxData []any
						json.Unmarshal([]byte(rawTx), &suiTxData)
						if suiTxData == nil || len(suiTxData) != 4 {
							return "", fmt.Errorf("invalid benfen tx data, recordId: %s", recordId)
						}
						res = suiTxData[1].([]any)[0].(string)
					}

				} else if action == "TRANSACTION_SIGNATURE" && appr.ExtraData.Authorization != nil {
					res = appr.ExtraData.Authorization.FinalHash
				}
				if res == "" {
					return "", fmt.Errorf("sign result is empty")
				}
				return res, nil

			case "REJECT":
				return "", fmt.Errorf("approval rejected")
			}
		}

		time.Sleep(2 * time.Second)
	}
	return "", fmt.Errorf("approve timeout")
}

func ConvertCompiledInstructions(instructions []solana.CompiledInstruction) []map[string]any {
	var convertedInstructions []map[string]any
	for _, instruction := range instructions {
		convertedInstruction := map[string]any{
			"programIdIndex":    instruction.ProgramIDIndex,
			"accountKeyIndexes": instruction.Accounts,
			"data":              base64.StdEncoding.EncodeToString(instruction.Data),
		}
		convertedInstructions = append(convertedInstructions, convertedInstruction)
	}
	return convertedInstructions
}
