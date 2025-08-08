package portal

import "testing"

type user struct {
	Email string `validate:"required,email"`
	Name  string `validate:"required"`
	Code  string `validate:"len=3"`
}

func TestValidator_PassesAndRejects(t *testing.T) {
	v := GetDefaultValidator()

	ok, err := v.Passes(&user{
		Email: "a@b.com",
		Name:  "John",
		Code:  "123",
	})

	if err != nil || !ok {
		t.Fatalf("expected pass got %v %v", ok, err)
	}

	invalid := &user{
		Email: "bad",
		Name:  "",
		Code:  "1",
	}

	if ok, err := v.Passes(invalid); ok || err == nil {
		t.Fatalf("expected fail")
	}

	if len(v.GetErrors()) == 0 {
		t.Fatalf("errors not recorded")
	}

	json := v.GetErrorsAsJson()

	if json == "" {
		t.Fatalf("json empty")
	}
}

func TestValidator_Rejects(t *testing.T) {
	v := GetDefaultValidator()
	u := &user{
		Email: "",
		Name:  "",
		Code:  "1",
	}

	reject, _ := v.Rejects(u)

	if !reject {
		t.Fatalf("expected reject")
	}
}
