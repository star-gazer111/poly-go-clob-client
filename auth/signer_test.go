package auth

import "testing"

func TestAPICredsStringIsRedacted(t *testing.T) {
	c := APICreds{
		Key:        "key_abcdefghijklmnopqrstuvwxyz",
		Secret:     "secret_abcdefghijklmnopqrstuvwxyz",
		Passphrase: "pass_abcdefghijklmnopqrstuvwxyz",
	}

	s := c.String()
	// Must not contain full raw values
	if contains(s, c.Key) || contains(s, c.Secret) || contains(s, c.Passphrase) {
		t.Fatalf("String() leaked secret: %s", s)
	}
}

func contains(hay, needle string) bool {
	return needle != "" && len(needle) > 0 && (len(hay) >= len(needle)) && (stringIndex(hay, needle) >= 0)
}

// small helper to avoid pulling strings package into tests? (we can just use strings.Contains)
func stringIndex(s, sub string) int {
	// naive index
	n := len(sub)
	for i := 0; i+n <= len(s); i++ {
		if s[i:i+n] == sub {
			return i
		}
	}
	return -1
}
