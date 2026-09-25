// Package seedjson はLocal開発用User seedのJSON入力を検証します。
package seedjson

import (
	"encoding/json"
	"fmt"
	"io"
	"net/mail"
	"net/url"
	"os"
	"strings"

	userdomain "github.com/iwasawa/hogedd-api/internal/user/domain"
	userpostgres "github.com/iwasawa/hogedd-api/internal/user/infrastructure/postgres"
)

type document struct {
	Version int    `json:"version"`
	Users   []user `json:"users"`
}

type user struct {
	AuthIssuer  string `json:"auth_issuer"`
	AuthSubject string `json:"auth_subject"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
	Status      string `json:"status"`
}

// Load は指定したJSONファイルを読み、検証済みのUser seedを返します。
func Load(path string) ([]userpostgres.LocalSeedUser, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open seed users file: %w", err)
	}
	defer file.Close()

	var input document
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		return nil, fmt.Errorf("decode seed users file: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return nil, fmt.Errorf("decode seed users file: trailing content")
	}
	if input.Version != 1 {
		return nil, fmt.Errorf("unsupported seed users version: %d", input.Version)
	}

	users := make([]userpostgres.LocalSeedUser, 0, len(input.Users))
	seen := make(map[string]struct{}, len(input.Users))
	for index, inputUser := range input.Users {
		parsed, err := parseUser(inputUser)
		if err != nil {
			return nil, fmt.Errorf("validate users[%d]: %w", index, err)
		}
		key := parsed.AuthIssuer + "\x00" + parsed.AuthSubject
		if _, exists := seen[key]; exists {
			return nil, fmt.Errorf("validate users[%d]: duplicate auth identity", index)
		}
		seen[key] = struct{}{}
		users = append(users, parsed)
	}
	return users, nil
}

func parseUser(input user) (userpostgres.LocalSeedUser, error) {
	issuer, err := url.ParseRequestURI(input.AuthIssuer)
	if err != nil || issuer.Scheme != "https" || issuer.Host == "" {
		return userpostgres.LocalSeedUser{}, fmt.Errorf("auth_issuer must be an absolute HTTPS URL")
	}
	if strings.TrimSpace(input.AuthSubject) == "" {
		return userpostgres.LocalSeedUser{}, fmt.Errorf("auth_subject is required")
	}
	address, err := mail.ParseAddress(input.Email)
	if err != nil || address.Address != input.Email {
		return userpostgres.LocalSeedUser{}, fmt.Errorf("email is invalid")
	}
	role, err := userdomain.ParseRole(input.Role)
	if err != nil {
		return userpostgres.LocalSeedUser{}, err
	}
	status, err := userdomain.ParseStatus(input.Status)
	if err != nil {
		return userpostgres.LocalSeedUser{}, err
	}
	profile, err := userdomain.NewProfile("seed-user", input.DisplayName)
	if err != nil {
		return userpostgres.LocalSeedUser{}, fmt.Errorf("display_name is invalid: %w", err)
	}
	return userpostgres.LocalSeedUser{
		AuthIssuer: issuer.String(), AuthSubject: input.AuthSubject, Email: input.Email,
		DisplayName: profile.DisplayName(), Role: role, Status: status,
	}, nil
}
