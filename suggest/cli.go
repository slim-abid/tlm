package suggest

import (
	"context"
	"errors"
	"fmt"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
	"github.com/yusufcanb/tlm/explain"
	"github.com/yusufcanb/tlm/shell"
	"os"
	"time"
)

func (s *Suggest) before(_ *cli.Context) error {
	_, err := s.api.Version(context.Background())
	if err != nil {
		fmt.Println(shell.Err() + " " + err.Error())
		fmt.Println(shell.Err() + " Ollama connection failed. Please check your Ollama if it's running or configured correctly.")
		os.Exit(-1)
	}

	return nil
}

func (s *Suggest) action(c *cli.Context) error {
	var responseText string
	var err error

	var t1, t2 time.Time

	prompt := c.Args().Get(0)
	spinner.New().
		Type(spinner.Line).
		Title(" Thinking...").
		Style(lipgloss.NewStyle().Foreground(lipgloss.Color("2"))).
		Action(func() {
			t1 = time.Now()
			responseText, err = s.getCommandSuggestionFor(Stable, viper.GetString("shell"), prompt)
			t2 = time.Now()
		}).
		Run()

	if err != nil {
		fmt.Println(shell.Err()+" error getting suggestion:", err)
	}

	fmt.Printf(shell.SuccessMessage("┃ >")+" Thinking... (%s)\n", t2.Sub(t1).String())
	form := NewCommandForm(s.extractCommandsFromResponse(responseText)[0])
	err = form.Run()

	fmt.Println(shell.SuccessMessage("┃ > ") + form.command)
	if err != nil {
		fmt.Println(shell.WarnMessage("┃ > ") + "Aborted..." + "\n")
		return nil
	}

	if form.action == Execute {
		fmt.Println(shell.SuccessMessage("┃ > ") + "Executing..." + "\n")
		cmd, stdout, stderr := shell.Exec2(form.command)
		err = cmd.Run()
		if err != nil {
			return err
		}

		if stderr.String() != "" {
			fmt.Println()
			return errors.New("command failed")
		}

		fmt.Println(stdout.String())
		return nil
	}

	if form.action == Explain {
		fmt.Println(shell.SuccessMessage("┃ > ") + "Explaining..." + "\n")

		exp := explain.New(s.api, "")
		err = exp.StreamExplanationFor(Stable, form.command)
		if err != nil {
			return err
		}

	} else {
		fmt.Println(shell.WarnMessage("┃ > ") + "Aborted..." + "\n")
	}

	return nil
}

func (s *Suggest) Command() *cli.Command {
	return &cli.Command{
		Name:    "suggest",
		Aliases: []string{"s"},
		Usage:   "suggest a command.",
		Before:  s.before,
		Action:  s.action,
	}
}
