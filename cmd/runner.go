package main

import (
	"flag"
	"log"
	"openblock-approval-sdk/approval"
	"time"
)

func main() {
	// 定义命令行参数
	configPath := flag.String("config", "config.json", "Path to the configuration file")
	flag.Parse()

	// 从配置文件加载参数
	wallet, err := approval.NewApprovalWalletFromJson(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration from %s: %v", *configPath, err)
	}

	for {
		switch wallet.Role {
		case "initiator":
			action := "TRANSACTION"
			if wallet.TxInfo.Msg != nil && wallet.TxInfo.Msg.SignMsg != "" {
				action = "TRANSACTION_SIGNATURE"
			}
			res, err := wallet.SendApprovalTxInfo(wallet.TxInfo.Chain, action, wallet.TxInfo)
			if err != nil {
				log.Fatalf("Approval fail: %v", err)
			}
			log.Printf("Approval response: %s", res)
			return

		case "approver":
			if err := wallet.AutoApprove(); err != nil {
				log.Fatalf("Auto approval failed: %v", err)
			}

		case "manager":
			if err := wallet.AutoSign(); err != nil {
				log.Fatalf("Auto sign failed: %v", err)
			}

		default:
			log.Fatalf("Unknown role: %s", wallet.Role)
			return
		}

		time.Sleep(5 * time.Second)
	}
}
