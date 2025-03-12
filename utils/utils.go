package utils

import (
	"crypto/rand"
	"encoding/base64"
	"html/template"
	"io"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// Wraps one template into another.
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
func PasswordAnalysis(password string) bool {
	const minLen, maxLen string = "8", "128"
	validPassword := regexp.MustCompile(`^[\x20-\x7E]{` + minLen + `,` + maxLen + `}$`)
	return validPassword.MatchString(password)
}

func UsernameAnalysis(username string) bool {
	const minLen, maxLen string = "3", "32"
	validUsername := regexp.MustCompile(`^[a-zA-Z0-9_.-]{` + minLen + `,` + maxLen + `}$`)
	return validUsername.MatchString(username)
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
