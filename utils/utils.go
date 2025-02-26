package utils

import (
	"crypto/rand"
	"encoding/base64"
	"html/template"
	"io"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// Wraps one template into another. In wrapper target should be named as {{ .Htmlstr }}, and wrapper data named Data
func TemplateWrap(tmpl *template.Template, w io.Writer, targetName string, targetData any, wrapperName string, wrapperData any) error {
	buff := &strings.Builder{}
	err := tmpl.ExecuteTemplate(buff, targetName, targetData)
	if err != nil {
		return err
	}

	wrapperCtx := &struct {
		Htmlstr template.HTML
		Data    any
	}{template.HTML(buff.String()), wrapperData}
	err = tmpl.ExecuteTemplate(w, wrapperName, wrapperCtx)
	if err != nil {
		return err
	}
	return nil
}

func HashPassword(password string) (string, error) {
	const cost int = 14
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	return string(bytes), err
}

// PasswordAnalysis returns if the password meets the requirements
func PasswordAnalysis(password string, minLength int, maxLength int) bool {
	var hasUpperCase, hasLoverCase, hasDigits, hasSpecialChars bool
	if len(password) < minLength {
		return false
	} else if len(password) > maxLength {
		return false
	}
	for _, char := range password {
		switch {
		case unicode.IsSpace(char):
			return false
		case !hasUpperCase && unicode.IsUpper(char):
			hasUpperCase = true
		case !hasLoverCase && unicode.IsLower(char):
			hasLoverCase = true
		case !hasDigits && unicode.IsDigit(char):
			hasDigits = true
		case !hasSpecialChars && unicode.IsPunct(char):
			hasSpecialChars = true
		}
		if hasUpperCase && hasDigits && hasSpecialChars && hasLoverCase {
			return true
		}
	}
	return false
}

// Generate secure token for session
func GenerateToken(n int) (string, error) {
	// n is the length of the token
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
