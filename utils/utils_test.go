package utils

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	type args struct {
		password string
	}
	tests := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{name: "Test 1", args: args{password: "password"}, want: nil, wantErr: false},
		{name: "Test 2", args: args{password: "random"}, want: nil, wantErr: false},
		{name: "Test 3", args: args{password: "BullShit8373"}, want: nil, wantErr: false},
		{name: "Test 4", args: args{password: "Whatever98367"}, want: nil, wantErr: false},
		{name: "Test 5", args: args{password: "Heroku88GGG"}, want: nil, wantErr: false},
		{name: "Test 6", args: args{password: "GolangIsTheBest"}, want: nil, wantErr: false},
		{name: "Test 7", args: args{password: "H_))eroku88G,,GG"}, want: nil, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got := bcrypt.CompareHashAndPassword([]byte(hash), []byte(tt.args.password))
			if got != tt.want {
				t.Errorf("HashPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkHashPassword(b *testing.B) {
	for range b.N {
		HashPassword("passadfasdfasdfasword")
	}
}

func TestPasswordAnalysis(t *testing.T) {
	type args struct {
		password  string
		minLength int
		maxLength int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "Test 1", args: args{password: "password", minLength: 8, maxLength: 128}, want: false},
		{name: "Test 2", args: args{password: "random", minLength: 8, maxLength: 128}, want: false},
		{name: "Test 3", args: args{password: "BullShit8373", minLength: 8, maxLength: 128}, want: false},
		{name: "Test 4", args: args{password: "Whatever98367_", minLength: 8, maxLength: 128}, want: true},
		{name: "Test 5", args: args{password: "H_))eroku88G,,GG", minLength: 8, maxLength: 128}, want: true},
		{name: "Test 6", args: args{password: "GolangIsTheBasdfalhsdfh;lasdlfas;dlkfaslk;dfas;dfasdfasdfa;svcamsd,fa,smdnfasdfasdjfalskdfaskdflkjasdkfjhasdhfasdfhjlasestasdfac39asdf2349798jksdfvskjl", minLength: 8, maxLength: 128}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PasswordAnalysis(tt.args.password, tt.args.minLength, tt.args.maxLength); got != tt.want {
				t.Errorf("PasswordAnalysis() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateToken(t *testing.T) {
	tests := []struct {
		name    string
		n       int
		want    string
		wantErr bool
	}{
		{name: "Test 1", n: 32, want: "", wantErr: false},
		{name: "Test 2", n: 64, want: "", wantErr: false},
		{name: "Test 3", n: 128, want: "", wantErr: false},
		{name: "Test 4", n: 256, want: "", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateToken(tt.n)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GenerateToken() = %v, want %v", got, tt.want)
			}
		})
	}
}
