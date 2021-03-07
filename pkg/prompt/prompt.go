package prompt

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
)

func Confirm(msg string) bool {
	prompt := &survey.Confirm{
		Message: msg,
		Default: false,
	}
	confirm := false
	err := survey.AskOne(prompt, &confirm)
	if err == terminal.InterruptErr {
		fmt.Println(err)
		os.Exit(1)
	}
	return confirm
}

func Confirmf(msg string, args ...interface{}) bool {
	return Confirm(fmt.Sprintf(msg, args...))
}
