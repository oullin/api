package auth_test

import (
	"reflect"
	"testing"

	"github.com/oullin/pkg/auth"
)

func TestSchemaConstants(t *testing.T) {
	tests := []struct {
		name string
		got  any
		want any
	}{
		{"PublicKeyPrefix", auth.PublicKeyPrefix, "pk_"},
		{"SecretKeyPrefix", auth.SecretKeyPrefix, "sk_"},
		{"TokenMinLength", auth.TokenMinLength, 16},
		{"AccountNameMinLength", auth.AccountNameMinLength, 5},
		{"EncryptionKeyLength", auth.EncryptionKeyLength, 32},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(tt.got, tt.want) {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.want)
			}
		})
	}
}
