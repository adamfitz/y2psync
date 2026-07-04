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
	"github.com/adam/y2psync/internal/scraper"
)

type PlaylistView struct {
	repo        *database.PlaylistRepo
	win         fyne.Window
	listSelect  *widget.Select
	entryList   *widget.List
	entries     []*model.PlaylistEntry
	importBtn   *widget.Button
	addVideoBtn *widget.Button
	createBtn   *widget.Button
	deleteBtn   *widget.Button
}

func NewPlaylistView(repo *database.PlaylistRepo, win fyne.Window) *PlaylistView {
	pv := &PlaylistView{repo: repo, win: win}

	pv.listSelect = widget.NewSelect(nil, func(selected string) {
		pv.loadEntries(selected)
	})

	pv.entryList = widget.NewList(
		func() int { return len(pv.entries) },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel(""),
				widget.NewLabel(""),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id >= len(pv.entries) {
				return
			}
			entry := pv.entries[id]
			box := item.(*fyne.Container)
			title := entry.YouTubeVideoID
			if entry.DisplayTitle != "" {
				title = entry.DisplayTitle
			}
			box.Objects[0].(*widget.Label).SetText(title)
			box.Objects[1].(*widget.Label).SetText(entry.CreatedAt.Format("2006-01-02 15:04"))
		},
	)

	pv.addVideoBtn = widget.NewButton("Add Video", func() {
		pv.showAddVideoDialog()
	})

	pv.importBtn = widget.NewButton("Import from YouTube URL", func() {
		pv.showImportDialog()
	})

	pv.createBtn = widget.NewButton("New Playlist", func() {
		pv.showCreateDialog()
	})

	pv.deleteBtn = widget.NewButton("Delete Playlist", func() {
		pv.deleteSelected()
	})

	pv.refreshList()

	return pv
}

func (pv *PlaylistView) Container() fyne.CanvasObject {
	topBar := container.NewHBox(
		widget.NewLabel("Playlist:"),
		pv.listSelect,
		pv.createBtn,
		pv.deleteBtn,
		pv.addVideoBtn,
		pv.importBtn,
	)

	split := container.NewBorder(topBar, nil, nil, nil, pv.entryList)
	return split
}

func (pv *PlaylistView) refreshList() {
	lists, err := pv.repo.ListLists()
	if err != nil {
		return
	}
	names := make([]string, 0, len(lists))
	for _, l := range lists {
		names = append(names, l.Name)
	}
	pv.listSelect.Options = names
	if len(names) > 0 {
		pv.listSelect.SetSelected(names[0])
	} else {
		pv.listSelect.ClearSelected()
		pv.entries = nil
		pv.entryList.Refresh()
	}
}

func (pv *PlaylistView) loadEntries(playlistName string) {
	if playlistName == "" {
		pv.entries = nil
		pv.entryList.Refresh()
		return
	}
	lists, _ := pv.repo.ListLists()
	var listID string
	for _, l := range lists {
		if l.Name == playlistName {
			listID = l.ID
			break
		}
	}
	if listID == "" {
		return
	}
	entries, err := pv.repo.GetEntries(listID)
	if err != nil {
		return
	}
	pv.entries = entries
	pv.entryList.Refresh()
}

func (pv *PlaylistView) showCreateDialog() {
	entry := widget.NewEntry()
	entry.SetPlaceHolder("Playlist name")

	dialog.ShowForm("New Playlist", "Create", "Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("Name", entry),
		},
		func(ok bool) {
			if !ok || entry.Text == "" {
				return
			}
			list := &model.PlaylistList{
				ID:        uuid.New().String(),
				Name:      entry.Text,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			}
			if err := pv.repo.CreateList(list); err != nil {
				dialog.ShowError(err, pv.win)
				return
			}
			pv.refreshList()
		},
		pv.win,
	)
}

func (pv *PlaylistView) deleteSelected() {
	selected := pv.listSelect.Selected
	if selected == "" {
		return
	}
	lists, _ := pv.repo.ListLists()
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
	if err := pv.repo.DeleteList(listID); err != nil {
		dialog.ShowError(err, pv.win)
		return
	}
	pv.refreshList()
}

func (pv *PlaylistView) showAddVideoDialog() {
	selected := pv.listSelect.Selected
	if selected == "" {
		dialog.ShowInformation("No Playlist", "Select a playlist first", pv.win)
		return
	}

	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("https://www.youtube.com/watch?v=... or https://youtu.be/...")

	dialog.ShowForm("Add Video to Playlist", "Add", "Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("Video URL", urlEntry),
		},
		func(ok bool) {
			if !ok || urlEntry.Text == "" {
				return
			}
			pv.addVideoToPlaylist(urlEntry.Text)
		},
		pv.win,
	)
}

func (pv *PlaylistView) addVideoToPlaylist(rawURL string) {
	videoID, err := extractVideoID(rawURL)
	if err != nil {
		dialog.ShowError(fmt.Errorf("invalid video URL: %w", err), pv.win)
		return
	}

	selected := pv.listSelect.Selected
	lists, _ := pv.repo.ListLists()
	var listID string
	for _, l := range lists {
		if l.Name == selected {
			listID = l.ID
			break
		}
	}
	if listID == "" {
		dialog.ShowError(fmt.Errorf("playlist not found"), pv.win)
		return
	}

	exists, err := pv.repo.EntryExists(listID, videoID)
	if err != nil {
		dialog.ShowError(err, pv.win)
		return
	}
	if exists {
		dialog.ShowInformation("Duplicate", "This video is already in the playlist", pv.win)
		return
	}

	maxOrder, err := pv.repo.GetMaxSortOrder(listID)
	if err != nil {
		maxOrder = 0
	}

	entry := &model.PlaylistEntry{
		ID:             uuid.New().String(),
		PlaylistListID: listID,
		YouTubeVideoID: videoID,
		DisplayTitle:   videoID,
		CreatedAt:      time.Now().UTC(),
		SortOrder:      maxOrder + 1,
	}
	if err := pv.repo.AddEntry(entry); err != nil {
		dialog.ShowError(err, pv.win)
		return
	}

	pv.loadEntries(selected)
}

func (pv *PlaylistView) showImportDialog() {
	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("https://www.youtube.com/playlist?list=...")

	targetSelect := widget.NewSelect(nil, nil)

	newNameEntry := widget.NewEntry()
	newNameEntry.SetPlaceHolder("New playlist name")

	useNewRadio := widget.NewRadioGroup([]string{"Existing", "New"}, func(selected string) {
		if selected == "Existing" {
			targetSelect.Show()
			newNameEntry.Hide()
		} else {
			targetSelect.Hide()
			newNameEntry.Show()
		}
	})
	useNewRadio.SetSelected("Existing")

	lists, _ := pv.repo.ListLists()
	names := make([]string, 0, len(lists))
	for _, l := range lists {
		names = append(names, l.Name)
	}
	targetSelect.Options = names
	if len(names) > 0 {
		targetSelect.SetSelected(names[0])
	}

	items := []*widget.FormItem{
		widget.NewFormItem("Playlist URL", urlEntry),
		widget.NewFormItem("Target", useNewRadio),
		widget.NewFormItem("Existing Playlist", targetSelect),
		widget.NewFormItem("New Playlist Name", newNameEntry),
	}

	targetSelect.Hide()

	dialog.ShowForm("Import YouTube Playlist", "Import", "Cancel", items,
		func(ok bool) {
			if !ok || urlEntry.Text == "" {
				return
			}
			pv.doImport(urlEntry.Text, useNewRadio.Selected, targetSelect.Selected, newNameEntry.Text)
		},
		pv.win,
	)
}

func (pv *PlaylistView) doImport(rawURL, targetType, existingName, newName string) {
	result, err := scraper.FetchPlaylistVideos(rawURL)
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to fetch playlist: %w", err), pv.win)
		return
	}

	var listID string
	if targetType == "Existing" {
		lists, _ := pv.repo.ListLists()
		for _, l := range lists {
			if l.Name == existingName {
				listID = l.ID
				break
			}
		}
	} else {
		list := &model.PlaylistList{
			ID:        uuid.New().String(),
			Name:      newName,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		if err := pv.repo.CreateList(list); err != nil {
			dialog.ShowError(err, pv.win)
			return
		}
		listID = list.ID
	}

	if listID == "" {
		dialog.ShowError(fmt.Errorf("no target playlist selected"), pv.win)
		return
	}

	maxOrder, err := pv.repo.GetMaxSortOrder(listID)
	if err != nil {
		maxOrder = 0
	}

	added := 0
	skipped := 0
	for _, v := range result.Videos {
		exists, err := pv.repo.EntryExists(listID, v.VideoID)
		if err != nil || exists {
			if exists {
				skipped++
			}
			continue
		}
		maxOrder++
		entry := &model.PlaylistEntry{
			ID:             uuid.New().String(),
			PlaylistListID: listID,
			YouTubeVideoID: v.VideoID,
			DisplayTitle:   v.Title,
			CreatedAt:      time.Now().UTC(),
			SortOrder:      maxOrder,
		}
		if err := pv.repo.AddEntry(entry); err != nil {
			dialog.ShowError(err, pv.win)
			return
		}
		added++
	}

	dialog.ShowInformation("Import Complete",
		fmt.Sprintf("Discovered: %d\nAdded: %d\nSkipped (duplicates): %d",
			len(result.Videos), added, skipped), pv.win)

	playlistName := existingName
	if targetType != "Existing" {
		playlistName = newName
	}
	pv.refreshList()
	pv.listSelect.SetSelected(playlistName)
}
