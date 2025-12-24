package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/OpenBlockResource/openblock-approval-sdk-go/approval"
)

func main() {
	// 定义命令行参数
	configPath := flag.String("config", "config.json", "Path to the configuration file")
	checkWallet := flag.String("check-wallet", "", "Check wallet information, e.g. -check-wallet=Solana,ETH ")
	hdWalletId := flag.String("hd-wallet-id", "", "ID of the HD wallet")
	flag.Parse()

	// 从配置文件加载参数
	wallet, err := approval.NewApprovalWalletFromJson(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration from %s: %v", *configPath, err)
	}
	if *checkWallet != "" {
		walletInfos, err := wallet.Client.GetWalletInfo()
		if err != nil {
			log.Fatalf("Failed to get wallet info: %v", err)
			return
		}
		for _, walletInfo := range *walletInfos {
			for chain, addr := range walletInfo.WalletAddressMap {
				if strings.Contains(*checkWallet, chain) {
					fmt.Printf("%s, %s, %s, %s\n", walletInfo.WalletName, walletInfo.WalletId, chain, addr)
				}
			}
		}
		return
	}

	for {
		switch wallet.Role {
		case "initiator":
			res, err := wallet.SendApprovalTxInfo(*hdWalletId, wallet.TxInfo)
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
