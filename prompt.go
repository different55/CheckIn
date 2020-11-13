package main

import (
	"errors"
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
	"os/user"
	"strings"
)

// PromptInput prompts the user for input and returns either a well-formed status (leading ~user) or an empty string.
func PromptInput() (string, error) {
	model, err := newInputModel()
	if err != nil {
		return "", err
	}
	p := tea.NewProgram(model)
	err = p.Start()
	if err != nil {
		return "", err
	}
	return model.String(), nil
}

type inputModel struct {
	// Contains the model of the child text entry widget
	input textinput.Model
	// The prompt presented to the user
	prompt string
	// Any error that occurs while processing.
	Err error
}

// newInputModel creates a new inputModel ready for use, populated with a random prompt.
func newInputModel() (*inputModel, error) {
	user, err := user.Current()
	if err != nil {
		return &inputModel{}, err
	}
	model := inputModel{}
	model.prompt = fmt.Sprintf("What's ~%s up to?", user.Username)

	model.input = textinput.NewModel()

	// The input's Prompt is used to prepend an unchangeable prefix to the input.
	model.input.Prompt = fmt.Sprintf("~%s", user.Username)
	model.input.SetValue(" ")
	model.input.CursorEnd()
	model.input.Focus()
	return &model, nil
}

// String will return either a valid status with all whitespace removed or a blank string.
func (m *inputModel) String() string {
	output := strings.TrimRight(m.input.Value(), " \n\t\r")
	if output == "" {
		return output
	}
	return fmt.Sprintf("%s%s", m.input.Prompt, output)
}

func (m *inputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *inputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.Err = errors.New("user cancelled input")
			fallthrough
		case tea.KeyCtrlD, tea.KeyEnter:
			return m, tea.Quit
		}

	case error:
		m.Err = msg
		return m, tea.Quit
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *inputModel) View() string {
	if termenv.ColorProfile() == termenv.Ascii {
		return fmt.Sprintf("%s %s", m.prompt, m.input.View())
	}
	return fmt.Sprintf("%s %s", termenv.String(m.prompt).Bold(), m.input.View())
}
