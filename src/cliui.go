package src

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type UI struct {
	*Group
	app        *tview.Application
	MsgInputs  chan string
	peerbox    *tview.TextView
	messagebox *tview.TextView
	inputbox   *tview.InputField
}

func NewUI(gr *Group) *UI {
	msgchan := make(chan string)
	app := tview.NewApplication()
	messagebox := tview.NewTextView().SetDynamicColors(true).SetChangedFunc(func() { app.Draw() })
	messagebox.SetBorder(true).SetBorderColor(tcell.ColorRed).SetTitle(gr.GroupName).SetTitleAlign(tview.AlignLeft).SetTitleColor(tcell.ColorWhite)
	input := tview.NewInputField().SetLabel(gr.UserName + "=").SetLabelColor(tcell.ColorRed).SetFieldWidth(0).SetFieldBackgroundColor(tcell.ColorBlack)
	input.SetBorder(true).SetBorderColor(tcell.ColorRed).SetTitle("Input").SetTitleAlign(tview.AlignLeft).SetTitleColor(tcell.ColorWhite).SetBorderPadding(0, 0, 1, 0)
	input.SetDoneFunc(func(key tcell.Key) {
		if key != tcell.KeyEnter {
			return
		}
		line := input.GetText()
		if len(line) == 0 {
			return
		}
		msgchan <- line
		input.SetText("")
	})

	peerbox := tview.NewTextView()
	peerbox.SetBorder(true).SetBorderColor(tcell.ColorRed).SetTitle("Peers").SetTitleAlign(tview.AlignLeft).SetTitleColor(tcell.ColorWhite)

	flex := tview.NewFlex().SetDirection(tview.FlexColumn).AddItem(messagebox, 0, 1, false).AddItem(peerbox, 0, 1, false).AddItem(input, 0, 1, true)
	app.SetRoot(flex, true)
	return &UI{
		Group:      gr,
		app:        app,
		messagebox: messagebox,
		inputbox:   input,
		peerbox:    peerbox,
		MsgInputs:  msgchan,
	}
}

func (ui *UI) Run() error {
	go ui.starteventhandler()
	go ui.InitializeGroup()
	defer ui.Close()
	return ui.app.Run()
}

// A method of UI that closes the UI app
func (ui *UI) Close() {
	ui.pscancel()
}

// A method of UI that handles UI events
func (ui *UI) starteventhandler() {
	refreshticker := time.NewTicker(time.Second)
	defer refreshticker.Stop()

	for {
		select {

		case msg := <-ui.MsgInputs:
			// Send the message to outbound queue
			ui.Outbound <- msg
			// Add the message to the message box as a self message
			ui.display_selfmessage(msg)

		case msg := <-ui.Inbound:
			// Print the recieved messages to the message box
			ui.display_chatmessage(msg)

		case <-refreshticker.C:
			// Refresh the list of peers in the chat room periodically
			ui.syncpeerbox()

		case <-ui.pscntx.Done():
			// End the event loop
			return
		}
	}
}

func (ui *UI) InitializeGroup() {
	oldGroup := ui.Group

	newGroup, err := JoinGroup(ui.Host, "Home")
	if err != nil {
		fmt.Fprintln(File, "Error while initilalizing the group")
	}
	fmt.Fprintln(File, "Successfully initialized to Home group")
	ui.announce_group("Initialized to home group")
	ui.Group = newGroup
	time.Sleep(5 * time.Second)
	oldGroup.Exit()
}
func (ui *UI) announce_group(msg string) {
	promt := fmt.Sprintln("[yellow]Announcement")
	fmt.Fprintln(ui.messagebox, promt)
}

// A method of UI that displays a message recieved from a peer
func (ui *UI) display_chatmessage(msg chatmessage) {
	prompt := fmt.Sprintf("[green]<%s>:[-]", msg.SenderName)
	fmt.Fprintf(ui.messagebox, "%s %s\n", prompt, msg.Message)
}

// A method of UI that displays a message recieved from self
func (ui *UI) display_selfmessage(msg string) {
	prompt := fmt.Sprintf("[blue]<%s>:[-]", ui.UserName)
	fmt.Fprintf(ui.messagebox, "%s %s\n", prompt, msg)
}
func (ui *UI) syncpeerbox() {
	// Retrieve the list of peers from the chatroom

	peers := ui.PeerList()
	// Clear() is not a threadsafe call
	// So we acquire the thread lock on it
	// ui.peerbox.Lock()

	// Clear the box
	ui.peerbox.Clear()
	// Release the lock
	// ui.peerbox.Unlock()
	// Iterate over the list of peers
	for _, p := range peers {
		// Generate the pretty version of the peer ID
		peerid := p.Pretty()
		// Shorten the peer ID
		peerid = peerid[len(peerid)-8:]
		// Add the peer ID to the peer box
		fmt.Fprintln(ui.peerbox, peerid)
	}

	// Refresh the UI
	ui.app.Draw()
}
