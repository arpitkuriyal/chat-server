package client

import (
	"fmt"
	"strings"

	"github.com/arpitkuriyal/chat-server/internal/common"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var commands = []string{
	"/nick",
	"/whoami",
	"/exit",
}

func colorForUser(name string) string {
	colors := []string{
		"red",
		"green",
		"yellow",
		"blue",
		"magenta",
		"cyan",
	}
	return colors[len(name)%len(colors)]
}

func StartTUI(c *Client) error {
	app := tview.NewApplication()

	messages := tview.NewTextView()
	messages.SetDynamicColors(true)
	messages.SetScrollable(true)
	messages.SetBorder(true)

	title := " Chat-Server "
	if c.IsHost {
		title = " Chat-Server (HOST) "
	}
	messages.SetTitle(title)

	usersView := tview.NewTextView()
	usersView.SetDynamicColors(true)
	usersView.SetBorder(true)
	usersView.SetTitle(" Active Users ")

	input := tview.NewInputField()
	input.SetLabel("You: ")
	input.SetFieldWidth(0)

	input.SetAutocompleteFunc(func(currentText string) []string {
		if !strings.HasPrefix(currentText, "/") {
			return nil
		}

		var matches []string
		for _, cmd := range commands {
			if strings.HasPrefix(cmd, currentText) {
				matches = append(matches, cmd)
			}
		}

		if len(matches) == 0 {
			return nil
		}
		return matches
	})

	input.SetAutocompletedFunc(func(text string, index, source int) bool {
		if source == tview.AutocompletedNavigate {
			return false
		}
		input.SetText(text + " ")
		return true
	})

	input.SetDoneFunc(func(key tcell.Key) {
		if key != tcell.KeyEnter {
			return
		}

		text := strings.TrimSpace(input.GetText())
		if text == "" {
			return
		}

		if err := c.SendMessage(text); err != nil {
			app.QueueUpdateDraw(func() {
				fmt.Fprintf(messages, "[red]send error: %v\n", err)
			})
			app.Stop()
			return
		}

		input.SetText("")
	})

	chatLayout := tview.NewFlex().
		AddItem(messages, 0, 3, false).
		AddItem(usersView, 25, 1, false)

	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(chatLayout, 0, 1, false).
		AddItem(input, 3, 0, true)

	go func() {
		for {
			var msg common.Message
			if err := c.Dec.Decode(&msg); err != nil {
				app.QueueUpdateDraw(func() {
					fmt.Fprintf(messages, "\n[red]server disconnected\n")
				})
				app.Stop()
				return
			}

			app.QueueUpdateDraw(func() {
				switch msg.Type {

				case "user-list":
					usersView.SetText("")
					for _, u := range msg.Users {
						color := colorForUser(u)
						if u == c.Username && c.IsHost {
							fmt.Fprintf(usersView, "[yellow::b]- %s (HOST)[-]\n", u)
						} else if u == c.Username {
							fmt.Fprintf(usersView, "[%s]- %s (you)[-]\n", color, u)
						} else {
							fmt.Fprintf(usersView, "[%s]- %s[-]\n", color, u)
						}
					}

				default:
					fmt.Fprintf(
						messages,
						"[yellow]%s:[white] %s\n",
						msg.From,
						msg.Text,
					)
				}
			})
		}
	}()

	return app.SetRoot(layout, true).
		EnableMouse(true).
		Run()
}
