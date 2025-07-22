package pkg

import (
	"fmt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"strings"
	"time"
	"unicode"
)

type Stringable struct {
	value string
}

func MakeStringable(value string) *Stringable {
	return &Stringable{
		value: strings.TrimSpace(value),
	}
}

func (s Stringable) ToLower() string {
	caser := cases.Lower(language.English)

	return strings.TrimSpace(caser.String(s.value))
}

func (s Stringable) ToSnakeCase() string {
	var result strings.Builder

	for i, r := range s.value {
		if unicode.IsUpper(r) {
			if i > 0 {
				result.WriteByte('_')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

func (s Stringable) Dd(abstract any) {
	fmt.Println(fmt.Sprintf("dd: %+v", abstract))
}

func (s Stringable) ToDatetime() (*time.Time, error) {
	parsed, err := time.Parse(time.DateOnly, s.value)

	if err != nil {
		return nil, fmt.Errorf("error parsing date string: %v", err)
	}

	now := time.Now()

	produce := time.Date(
		parsed.Year(),
		parsed.Month(),
		parsed.Day(),
		now.Hour(),
		now.Minute(),
		now.Second(),
		now.Nanosecond(),
		now.Location(),
	)

	return &produce, nil
}
