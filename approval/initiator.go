package approval

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	apisdk "github.com/OpenBlockResource/openblock-api-sdk-go"
	solana "github.com/gagliardetto/solana-go"
)

const SOLANA = "Solana"

func SendApprovalTransaction(client *Client, hdWalletId, chainName, txData string) (string, error) {
	var txInfo *apisdk.TXInfo
	switch chainName {
	case SOLANA:
		transaction, err := solana.TransactionFromBase64(txData)
		if err != nil {
			return "", err
		}
		txInput := map[string]interface{}{
			"recent_blockhash": transaction.Message.RecentBlockhash.String(),
			"header": map[string]interface{}{
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
			TxPayload: []interface{}{
				txInput,
			},
		}

	default:
		return "", fmt.Errorf("not supported")
	}

	return SendApprovalTxInfo(client, hdWalletId, "TRANSACTION", txInfo)
}

func SignApprovalMessage(client *Client, hdWalletId, chainName, message string) (string, error) {
	var txInfo *apisdk.TXInfo
	switch chainName {
	case SOLANA:
		m, _ := hex.DecodeString(message)

		txInfo = &apisdk.TXInfo{
			Msg: &apisdk.Msg{
				SignMsg: message,
				Message: string(m),
			},
		}

	default:
		return "", fmt.Errorf("not supported")
	}

	return SendApprovalTxInfo(client, hdWalletId, "TRANSACTION_SIGNATURE", txInfo)
}

func SendApprovalTxInfo(client *Client, hdWalletId, action string, txInfo *apisdk.TXInfo) (string, error) {
	appr, err := client.NewApproval(hdWalletId, action, txInfo, "")
	if err != nil {
		return "", err
	}
	recordId := appr.Data.OriginRecordId

	for i := 0; i < 15; i++ {
		apprs, _ := client.GetSponsoredApprovals(recordId)
		for _, appr := range apprs.Data.Data {
			if appr.RecordID != recordId {
				continue
			}

			switch appr.Status {
			case "AGREE":
				res := appr.TxHash
				if action == "TRANSACTION_SIGNATURE" && appr.ExtraData.Authorization != nil {
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

func ConvertCompiledInstructions(instructions []solana.CompiledInstruction) []map[string]interface{} {
	var convertedInstructions []map[string]interface{}
	for _, instruction := range instructions {
		convertedInstruction := map[string]interface{}{
			"programIdIndex":    instruction.ProgramIDIndex,
			"accountKeyIndexes": instruction.Accounts,
			"data":              base64.StdEncoding.EncodeToString(instruction.Data),
		}
		convertedInstructions = append(convertedInstructions, convertedInstruction)
	}
	return convertedInstructions
}
