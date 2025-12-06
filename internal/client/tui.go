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

	// chat window
	messages := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() {
			app.Draw()
		})

	title := " Chat-Server "
	if c.IsHost {
		title = " Chat-Server (HOST) "
	}
	messages.SetBorder(true).SetTitle(title)

	// input box
	input := tview.NewInputField().
		SetLabel("You: ").
		SetFieldWidth(0)

	input.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			text := strings.TrimSpace(input.GetText())
			if text == "" {
				return
			}

			// later you can parse commands here if c.IsHost (like /kick user)
			if err := c.SendMessage(text); err != nil {
				app.QueueUpdateDraw(func() {
					fmt.Fprintf(messages, "[red]send error: %v\n", err)
				})
				app.Stop()
				return
			}

			input.SetText("")
		}
	})

	// layout: messages on top, input at bottom
	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(messages, 0, 1, false).
		AddItem(input, 3, 0, true)

	// goroutine: read messages from server
	go func() {
		for {
			var msg common.Message
			if err := c.Dec.Decode(&msg); err != nil {
				app.QueueUpdateDraw(func() {
					fmt.Fprintf(messages, "\n[red]server disconnected: %v\n", err)
				})
				return
			}

			app.QueueUpdateDraw(func() {
				fmt.Fprintf(messages, "[yellow]%s:[white] %s\n", msg.From, msg.Text)
			})
		}
	}()

	// run the TUI
	return app.SetRoot(layout, true).EnableMouse(true).Run()
}
