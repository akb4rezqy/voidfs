package auth

import (
	"errors"
	"strings"

	pam "github.com/msteinert/pam/v2"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type Authenticator interface {
	Authenticate(username, password string) error
}

type PAMAuthenticator struct {
	Service     string
	AllowedUser string
}

func NewPAMAuthenticator(service, allowedUser string) *PAMAuthenticator {
	if service == "" {
		service = "voidfs"
	}
	return &PAMAuthenticator{Service: service, AllowedUser: allowedUser}
}

func (a *PAMAuthenticator) Authenticate(username, password string) error {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return ErrInvalidCredentials
	}
	if a.AllowedUser != "" && username != a.AllowedUser {
		return ErrInvalidCredentials
	}

	tx, err := pam.StartFunc(a.Service, username, func(style pam.Style, message string) (string, error) {
		switch style {
		case pam.PromptEchoOff:
			return password, nil
		case pam.PromptEchoOn:
			return username, nil
		case pam.ErrorMsg, pam.TextInfo:
			return "", nil
		default:
			return "", ErrInvalidCredentials
		}
	})
	if err != nil {
		return ErrInvalidCredentials
	}
	if err := tx.Authenticate(0); err != nil {
		return ErrInvalidCredentials
	}
	if err := tx.AcctMgmt(0); err != nil {
		return ErrInvalidCredentials
	}
	return nil
}
