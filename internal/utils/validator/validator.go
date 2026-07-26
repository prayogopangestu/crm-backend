package validator

import (
	"fmt"
	"regexp"
	"strings"
)

type ValidationErrors map[string]string

func (e ValidationErrors) Error() string {
	return fmt.Sprintf("validation failed: %v", map[string]string(e))
}

func Required(value, field string) ValidationErrors {
	if strings.TrimSpace(value) == "" {
		return ValidationErrors{field: field + " wajib diisi"}
	}
	return nil
}

func Email(value, field string) ValidationErrors {
	pattern := `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
	if !regexp.MustCompile(pattern).MatchString(value) {
		return ValidationErrors{field: field + " tidak valid"}
	}
	return nil
}

func MinLength(value, field string, min int) ValidationErrors {
	if len([]rune(value)) < min {
		return ValidationErrors{field: field + " minimal " + fmt.Sprintf("%d", min) + " karakter"}
	}
	return nil
}

func Merge(errs ...ValidationErrors) ValidationErrors {
	merged := ValidationErrors{}
	for _, e := range errs {
		for k, v := range e {
			merged[k] = v
		}
	}
	if len(merged) == 0 {
		return nil
	}
	return merged
}
