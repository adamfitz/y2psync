package ui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/google/uuid"

	"github.com/adam/y2psync/internal/database"
	"github.com/adam/y2psync/internal/model"
)

type SubscriptionView struct {
	repo       *database.SubscriptionRepo
	win        fyne.Window
	entryList  *widget.List
	entries    []*model.SubscriptionEntry
	addBtn     *widget.Button
	deleteBtn  *widget.Button
	listID     string
	selected   int
	countLabel *widget.Label
}

func NewSubscriptionView(repo *database.SubscriptionRepo, win fyne.Window) *SubscriptionView {
	sv := &SubscriptionView{repo: repo, win: win, selected: -1}

	sv.entryList = widget.NewList(
		func() int { return len(sv.entries) },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel(""),
				widget.NewLabel(""),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id >= len(sv.entries) {
				return
			}
			entry := sv.entries[id]
			box := item.(*fyne.Container)
			name := entry.YouTubeChannelID
			if entry.ChannelName != "" {
				name = entry.ChannelName
			}
			box.Objects[0].(*widget.Label).SetText(name)
			box.Objects[1].(*widget.Label).SetText(entry.CreatedAt.Format("2006-01-02 15:04"))
		},
	)

	sv.entryList.OnSelected = func(id widget.ListItemID) {
		sv.selected = int(id)
	}

	sv.addBtn = widget.NewButton("Add Channel URL", func() {
		sv.showAddDialog()
	})

	sv.deleteBtn = widget.NewButton("Delete Channel", func() {
		sv.deleteSelected()
	})

	sv.countLabel = widget.NewLabel("")

	sv.ensureDefaultList()
	sv.loadEntries()

	return sv
}

func (sv *SubscriptionView) Container() fyne.CanvasObject {
	topBar := container.NewHBox(
		widget.NewLabel("Subscriptions"),
		sv.addBtn,
		sv.deleteBtn,
	)

	split := container.NewBorder(topBar, sv.countLabel, nil, nil, sv.entryList)
	return split
}

func (sv *SubscriptionView) ensureDefaultList() {
	lists, err := sv.repo.ListLists()
	if err != nil {
		return
	}

	defaultName := "My Subscriptions"
	for _, l := range lists {
		if l.Name == defaultName {
			sv.listID = l.ID
			return
		}
	}

	list := &model.SubscriptionList{
		ID:        uuid.New().String(),
		Name:      defaultName,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := sv.repo.CreateList(list); err != nil {
		return
	}
	sv.listID = list.ID
}

func (sv *SubscriptionView) refreshEntries() {
	sv.loadEntries()
}

func (sv *SubscriptionView) loadEntries() {
	if sv.listID == "" {
		sv.countLabel.SetText("")
		return
	}
	entries, err := sv.repo.GetEntries(sv.listID)
	if err != nil {
		return
	}
	sv.entries = entries
	sv.entryList.Refresh()
	sv.countLabel.SetText(fmt.Sprintf("%d channels", len(entries)))
}

func (sv *SubscriptionView) showAddDialog() {
	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("https://www.youtube.com/channel/UC... or @handle")

	dialog.ShowForm("Add Channel Subscription", "Add", "Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("Channel URL", urlEntry),
		},
		func(ok bool) {
			if !ok || urlEntry.Text == "" {
				return
			}
			sv.addChannel(urlEntry.Text)
		},
		sv.win,
	)
}

func (sv *SubscriptionView) addChannel(rawURL string) {
	channelID, err := extractChannelID(rawURL)
	if err != nil {
		dialog.ShowError(err, sv.win)
		return
	}

	exists, err := sv.repo.EntryExists(sv.listID, channelID)
	if err != nil {
		dialog.ShowError(err, sv.win)
		return
	}
	if exists {
		dialog.ShowInformation("Duplicate", "Channel is already subscribed", sv.win)
		return
	}

	entry := &model.SubscriptionEntry{
		ID:                 uuid.New().String(),
		SubscriptionListID: sv.listID,
		YouTubeChannelID:   channelID,
		ChannelURL:         rawURL,
		CreatedAt:          time.Now().UTC(),
	}
	if err := sv.repo.AddEntry(entry); err != nil {
		dialog.ShowError(err, sv.win)
		return
	}

	sv.loadEntries()
}

func (sv *SubscriptionView) deleteSelected() {
	if sv.selected < 0 || sv.selected >= len(sv.entries) {
		dialog.ShowInformation("No Selection", "Click a subscription in the list to select it, then press Delete", sv.win)
		return
	}
	entry := sv.entries[sv.selected]
	dialog.ShowConfirm("Remove Subscription",
		fmt.Sprintf("Remove %s from subscriptions?", entry.YouTubeChannelID),
		func(ok bool) {
			if !ok {
				return
			}
			if err := sv.repo.RemoveEntry(entry.ID); err != nil {
				dialog.ShowError(err, sv.win)
				return
			}
			sv.selected = -1
			sv.loadEntries()
		}, sv.win)
}
