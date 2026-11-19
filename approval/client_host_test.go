package approval

import "testing"

func TestNewClientUsesConfiguredAPIHost(t *testing.T) {
	const host = "mpc.example.com"

	client := NewClient("key", "secret", host)
	if got := client.apiClient.CompanyWallet.Host; got != host {
		t.Fatalf("API host = %q, want %q", got, host)
	}
}

func TestNewClientKeepsDefaultAPIHostWhenUnset(t *testing.T) {
	client := NewClient("key", "secret")
	if got := client.apiClient.CompanyWallet.Host; got != "auth.openblock.com" {
		t.Fatalf("API host = %q, want default auth.openblock.com", got)
	}
}

func TestNewApprovalWalletPassesConfiguredAPIHost(t *testing.T) {
	const host = "mpc.example.com"

	wallet := NewApprovalWallet("key", "secret", host)
	if got := wallet.Client.apiClient.CompanyWallet.Host; got != host {
		t.Fatalf("API host = %q, want %q", got, host)
	}
}
