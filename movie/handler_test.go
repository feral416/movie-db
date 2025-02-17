package movie

import (
	"net/http"
	"strings"
	"testing"
)

func TestUserRegistration(t *testing.T) {
	client := &http.Client{}
	type args struct {
		username string
		password string
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "Test 1", args: args{username: "user1", password: "H3roj__demco"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client.Post("http://localhost:8080/user/register", "application/x-www-form-urlencoded", strings.NewReader("username="+tt.args.username+"&passweord="+tt.args.password))
		})
	}
}
