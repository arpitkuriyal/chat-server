package client

import (
	"fmt"
	"strings"

	"github.com/arpitkuriyal/chat-server/internal/common"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// StartTUI builds and runs the tview UI for this client.
func StartTUI(c *Client) error {
	app := tview.NewApplication()
	messages := tview.NewTextView()
	messages.
		SetDynamicColors(true).
		SetScrollable(true).
		SetBorder(true)

	title := " Chat-Server "
	if c.IsHost {
		title = " Chat-Server (HOST) "
	}
	messages.SetTitle(title)

	usersView := tview.NewTextView()
	usersView.
		SetBorder(true).
		SetTitle(" Active Users ")

	input := tview.NewInputField()
	input.
		SetLabel("You: ").
		SetFieldWidth(0)

	input.SetDoneFunc(func(key tcell.Key) {
		if key != tcell.KeyEnter {
			return
		}

		text := strings.TrimSpace(input.GetText())
		if text == "" {
			return
		}

		if text == "/exit" {
			_ = c.SendMessage(text)
			app.Stop()
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
						if u == c.Username {
							fmt.Fprintf(usersView, "[green]- %s (you)\n", u)
						} else {
							fmt.Fprintf(usersView, "- %s\n", u)
						}
					}

				default:
					fmt.Fprintf(messages, "[yellow]%s:[white] %s\n", msg.From, msg.Text)
				}
			})
		}
	}()

	return app.SetRoot(layout, true).
		EnableMouse(true).
		Run()
}
