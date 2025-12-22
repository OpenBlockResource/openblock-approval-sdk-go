package approval

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type SignResult struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func AutoSign(client *Client, approvalParams *[]ApprovalParams) error {
	result, err := AutoApprove(client, approvalParams)
	if err != nil {
		return err
	}
	for _, res := range result {
		if !res.Approved {
			continue
		}
		url := "http://localhost:7790/openapi/sign/%s?key=%s"
		if res.Action == "TRANSACTION_SIGNATURE" {
			url = fmt.Sprintf(url, "sign_message", client.ApiKey)
		} else if res.OnlySign {
			url = fmt.Sprintf(url, "sign_transaction", client.ApiKey)
		} else {
			url = fmt.Sprintf(url, "send_transaction", client.ApiKey)
		}

		data := fmt.Sprintf(`{"company_wallet_approve_record_id": "%s"}`, res.ApprovalId)
		resp, err := http.Post(url, "application/json", bytes.NewBufferString(data))
		if err != nil {
			return fmt.Errorf("failed to send sign request: %w", err)
		}
		defer resp.Body.Close()

		// Read response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		// Check HTTP status code
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("sign request failed with status code: %d, body: %s", resp.StatusCode, string(body))
		}

		var signRes SignResult
		err = json.Unmarshal(body, &signRes)
		if err != nil {
			return fmt.Errorf("failed to unmarshal response: %w, body: %s", err, string(body))
		}

		if signRes.Code != 0 {
			return errors.New(signRes.Message)
		}
		fmt.Printf("Approval ID %s signed successfully, result: %s\n", res.ApprovalId, signRes.Data)
	}
	return nil
}
