package auth

import "github.com/msteinert/pam/v2"

func Authenticate(username, password string) error {
	
	t, err := pam.StartFunc("gocockpit", username, func(
		style pam.Style,
		msg string,
	) (string, error) {

		switch style {
		case pam.PromptEchoOn:
			return username, nil

		case pam.PromptEchoOff:
			return password, nil

		default:
			return "", nil
		}
	})

	if err != nil {
		return err
	}

	defer t.End()

	return t.Authenticate(0)
}