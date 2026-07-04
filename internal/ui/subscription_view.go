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
	listSelect *widget.Select
	entryList  *widget.List
	entries    []*model.SubscriptionEntry
	addBtn     *widget.Button
	createBtn  *widget.Button
	deleteBtn  *widget.Button
}

func NewSubscriptionView(repo *database.SubscriptionRepo, win fyne.Window) *SubscriptionView {
	sv := &SubscriptionView{repo: repo, win: win}

	sv.listSelect = widget.NewSelect(nil, func(selected string) {
		sv.loadEntries(selected)
	})

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

	sv.addBtn = widget.NewButton("Add Channel URL", func() {
		sv.showAddDialog()
	})

	sv.createBtn = widget.NewButton("New List", func() {
		sv.showCreateDialog()
	})

	sv.deleteBtn = widget.NewButton("Delete List", func() {
		sv.deleteSelected()
	})

	sv.refreshList()

	return sv
}

func (sv *SubscriptionView) Container() fyne.CanvasObject {
	topBar := container.NewHBox(
		widget.NewLabel("List:"),
		sv.listSelect,
		sv.createBtn,
		sv.deleteBtn,
		sv.addBtn,
	)

	split := container.NewBorder(topBar, nil, nil, nil, sv.entryList)
	return split
}

func (sv *SubscriptionView) refreshList() {
	lists, err := sv.repo.ListLists()
	if err != nil {
		return
	}
	names := make([]string, 0, len(lists))
	for _, l := range lists {
		names = append(names, l.Name)
	}
	sv.listSelect.Options = names
	if len(names) > 0 {
		sv.listSelect.SetSelected(names[0])
	} else {
		sv.listSelect.ClearSelected()
		sv.entries = nil
		sv.entryList.Refresh()
	}
}

func (sv *SubscriptionView) loadEntries(listName string) {
	if listName == "" {
		sv.entries = nil
		sv.entryList.Refresh()
		return
	}
	lists, _ := sv.repo.ListLists()
	var listID string
	for _, l := range lists {
		if l.Name == listName {
			listID = l.ID
			break
		}
	}
	if listID == "" {
		return
	}
	entries, err := sv.repo.GetEntries(listID)
	if err != nil {
		return
	}
	sv.entries = entries
	sv.entryList.Refresh()
}

func (sv *SubscriptionView) showCreateDialog() {
	entry := widget.NewEntry()
	entry.SetPlaceHolder("List name")

	dialog.ShowForm("New Subscription List", "Create", "Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("Name", entry),
		},
		func(ok bool) {
			if !ok || entry.Text == "" {
				return
			}
			list := &model.SubscriptionList{
				ID:        uuid.New().String(),
				Name:      entry.Text,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			}
			if err := sv.repo.CreateList(list); err != nil {
				dialog.ShowError(err, sv.win)
				return
			}
			sv.refreshList()
		},
		sv.win,
	)
}

func (sv *SubscriptionView) deleteSelected() {
	selected := sv.listSelect.Selected
	if selected == "" {
		return
	}
	lists, _ := sv.repo.ListLists()
	var listID string
	for _, l := range lists {
		if l.Name == selected {
			listID = l.ID
			break
		}
	}
	if listID == "" {
		return
	}
	if err := sv.repo.DeleteList(listID); err != nil {
		dialog.ShowError(err, sv.win)
		return
	}
	sv.refreshList()
}

func (sv *SubscriptionView) showAddDialog() {
	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("https://www.youtube.com/channel/UC... or @handle")

	dialog.ShowForm("Add Channel", "Add", "Cancel",
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
	selected := sv.listSelect.Selected
	if selected == "" {
		dialog.ShowError(fmt.Errorf("no subscription list selected"), sv.win)
		return
	}

	lists, _ := sv.repo.ListLists()
	var listID string
	for _, l := range lists {
		if l.Name == selected {
			listID = l.ID
			break
		}
	}
	if listID == "" {
		return
	}

	channelID, err := extractChannelID(rawURL)
	if err != nil {
		dialog.ShowError(err, sv.win)
		return
	}

	exists, err := sv.repo.EntryExists(listID, channelID)
	if err != nil {
		dialog.ShowError(err, sv.win)
		return
	}
	if exists {
		dialog.ShowInformation("Duplicate", "Channel is already in this list", sv.win)
		return
	}

	entry := &model.SubscriptionEntry{
		ID:               uuid.New().String(),
		SubscriptionListID: listID,
		YouTubeChannelID: channelID,
		ChannelURL:       rawURL,
		CreatedAt:        time.Now().UTC(),
	}
	if err := sv.repo.AddEntry(entry); err != nil {
		dialog.ShowError(err, sv.win)
		return
	}

	dialog.ShowInformation("Added", "Channel added to list", sv.win)
	sv.loadEntries(selected)
}
