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
		SetScrollable(true)

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

	// layout
	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(messages, 0, 1, false).
		AddItem(input, 3, 0, true)

	// 🔑 START THE UI EVENT LOOP FIRST
	go func() {
		if err := app.SetRoot(layout, true).EnableMouse(true).Run(); err != nil {
			panic(err)
		}
	}()

	// 🔑 NOW start reading messages from the server
	go func() {
		for {
			var msg common.Message
			if err := c.Dec.Decode(&msg); err != nil {
				app.QueueUpdateDraw(func() {
					fmt.Fprintf(messages, "\n[red]server disconnected\n")
				})
				return
			}

			app.QueueUpdateDraw(func() {
				fmt.Fprintf(messages, "[yellow]%s:[white] %s\n", msg.From, msg.Text)
			})
		}
	}()

	// block forever (tview owns the lifecycle now)
	select {}
}
