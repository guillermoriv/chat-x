package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.org/x/term"
)

func main() {
	if !term.IsTerminal(int(syscall.Stdin)) {
		fmt.Println("❌ This program must be run from a terminal.")
		os.Exit(1)
	}

	conn, err := net.Dial("tcp", "192.168.1.213:9000") // or server IP
	if err != nil {
		log.Println("Could not connect:", err)
		return
	}
	defer conn.Close()

	app := tview.NewApplication()

	// Chat log view
	chatView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() { app.Draw() })
	chatView.SetBorder(true).SetTitle(" Chat-X ")

	// User list view
	userListView := tview.NewTextView().
		SetDynamicColors(true).
		SetChangedFunc(func() { app.Draw() })
	userListView.SetBorder(true).SetTitle(" Users ")

	// Input field
	var inputField *tview.InputField
	inputField = tview.NewInputField().
		SetPlaceholder("Type your message here...").
		SetLabel("> ").
		SetFieldWidth(0).
		SetDoneFunc(func(key tcell.Key) {
			text := inputField.GetText()
			if text == "!quit" || text == "!exit" {
				app.Stop()
				return
			}
			fmt.Fprintln(conn, text)
			inputField.SetText("") // Clear after sending
		})
	inputField.SetBorder(true).SetTitle(" Message ")

	// Vertical layout for chat + input
	mainColumn := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(chatView, 0, 1, false).
		AddItem(inputField, 3, 0, true)

	// Main layout with user list to the right
	layout := tview.NewFlex().
		AddItem(mainColumn, 0, 4, true).
		AddItem(userListView, 25, 1, false)

	// Connection read loop
	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			line := scanner.Text()

			// Handle user list updates
			if strings.HasPrefix(line, "!users: ") {
				users := strings.TrimPrefix(line, "!users: ")
				app.QueueUpdateDraw(func() {
					userListView.Clear()
					userList := strings.Split(users, ",")

					for _, name := range userList {
						fmt.Fprintf(userListView, "%s\n", strings.TrimSpace(name))
					}
				})

				continue
			}

			// Regular chat message
			app.QueueUpdateDraw(func() {
				fmt.Fprintf(chatView, "%s\n", line)
			})
		}

		// On disconnect, show countdown
		app.QueueUpdateDraw(func() {
			fmt.Fprintln(chatView, "[red] Disconnected from server.")
		})

		countdown := 5
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			app.QueueUpdateDraw(func() {
				fmt.Fprintf(chatView, "[red] Closing in %d seconds...\n", countdown)
			})
			countdown--
			if countdown == 0 {
				app.Stop()
				return
			}
		}
	}()

	// run the UI
	if err := app.SetRoot(layout, true).EnableMouse(true).Run(); err != nil {
		log.Fatal(err)
	}
}
