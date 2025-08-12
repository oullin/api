package auth

import "testing"

func TestSchemaConstants(t *testing.T) {
	tests := []struct {
		name string
		got  any
		want any
	}{
		{"PublicKeyPrefix", PublicKeyPrefix, "pk_"},
		{"SecretKeyPrefix", SecretKeyPrefix, "sk_"},
		{"TokenMinLength", TokenMinLength, 16},
		{"AccountNameMinLength", AccountNameMinLength, 5},
		{"EncryptionKeyLength", EncryptionKeyLength, 32},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.want)
			}
		})
	}
}
