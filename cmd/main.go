package main

import (
	"fmt"
	"image/color"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("File Browser")

	myWindow.Resize(fyne.NewSize(800, 600))

	input := widget.NewEntry()
	input.SetPlaceHolder("Enter a directory path")

	searchInput := widget.NewEntry()
	searchInput.SetPlaceHolder("Search by filename")

	progressBar := widget.NewProgressBarInfinite()
	progressContainer := container.NewMax(progressBar)
	progressContainer.Hide()

	back := widget.NewButton("Back", nil)
	backContainer := container.NewHBox(layout.NewSpacer(), back, layout.NewSpacer())

	top := container.NewVBox(input, searchInput, progressContainer, backContainer)

	fileList := container.NewVBox()
	scrollContainer := container.NewScroll(fileList)

	content := container.NewBorder(top, nil, nil, nil, scrollContainer)

	myWindow.SetContent(content)

	back.OnTapped = func() {
		dir := filepath.Dir(input.Text)
		input.SetText(dir)
		browseDir(dir, fileList, myWindow)
	}

	input.OnSubmitted = func(text string) {
		browseDir(text, fileList, myWindow)
	}

	searchInput.OnSubmitted = func(text string) {
		go searchDir(input.Text, text, fileList, progressContainer)
	}

	myWindow.ShowAndRun()
}

func browseDir(dir string, fileList *fyne.Container, myWindow fyne.Window) {
	list, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	fileList.Objects = nil

	for _, item := range list {
		item := item
		name := item.Name()
		icon := theme.DocumentIcon()
		if item.IsDir() {
			icon = theme.FolderOpenIcon()
		}
		button := widget.NewButtonWithIcon(name, icon, func() {
			if item.IsDir() {
				newDir := filepath.Join(dir, name)
				browseDir(newDir, fileList, myWindow)
			} else {
				fileDialog := widget.NewLabel("File: " + name + "\nPath: " + dir)
				confirmButton := widget.NewButton("Open", func() {
					openFile(dir, name)
				})
				cancelButton := widget.NewButton("Delete", func() {
					deleteFile(dir, name)
					browseDir(dir, fileList, myWindow)
				})

				dialogContent := container.NewVBox(fileDialog, confirmButton, cancelButton)
				dialog := dialog.NewCustom("File Dialog", "Close", dialogContent, myWindow)
				dialog.Show()

				dialog.SetOnClosed(func() {
				})

			}
		})
		fileList.Add(button)
	}

	fileList.Refresh()
}

func deleteFile(dir string, name string) {
	err := os.Remove(filepath.Join(dir, name))
	if err != nil {
		panic(err)
	}
}

func openFile(dir string, name string) {
	fullPath := filepath.Join(dir, name)

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", fullPath)
	case "darwin":
		cmd = exec.Command("open", fullPath)
	case "linux":
		cmd = exec.Command("xdg-open", fullPath)
	default:
		fmt.Printf("unsupported operating system: %s\n", runtime.GOOS)
		return
	}

	err := cmd.Start()
	if err != nil {
		fmt.Printf("error opening file: %v\n", err)
	}
}


func searchDir(dir string, searchText string, fileList *fyne.Container, progressContainer *fyne.Container) {
    progressContainer.Show()

    fileList.Objects = nil

    var wg sync.WaitGroup
    fileChan := make(chan string)

    go func() {
        for f := range fileChan {
            name := filepath.Base(f)
            icon := theme.DocumentIcon()
            dirPath := filepath.Dir(f)
            if filepath.Dir(f) == dir {
                icon = theme.FolderOpenIcon()
            }
            fileButton := widget.NewButtonWithIcon(name, icon, func() {})
            dirLabel := canvas.NewText(dirPath, color.White)
            dirLabel.TextSize = 10 // set the text size to 10 (you can adjust this value to your liking)
            dirContainer := container.NewHBox(layout.NewSpacer(), dirLabel, layout.NewSpacer())
			buttonAndDir := container.NewVBox(layout.NewSpacer(), fileButton, layout.NewSpacer(), dirContainer)
            singleContainer := container.NewVBox(buttonAndDir)
            fileList.Add(singleContainer)
            fileList.Refresh()
        }
    }()

    // Assume dir is root directory
    rootDir, err := ioutil.ReadDir(dir)
    if err != nil {
        fmt.Printf("error reading the directory %v: %v\n", dir, err)
    }

    for _, d := range rootDir {
        if d.IsDir() {
            wg.Add(1)
            go func(d os.FileInfo) {
                defer wg.Done()
                err := filepath.Walk(filepath.Join(dir, d.Name()), func(path string, info os.FileInfo, err error) error {
                    if err != nil {
                        if os.IsPermission(err) {
                            return filepath.SkipDir
                        }
                        fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
                        return err
                    }

                    if info.IsDir() {
                        return nil
                    }

                    if strings.Contains(strings.ToLower(info.Name()), strings.ToLower(searchText)) {
                        fileChan <- path
                    }

                    return nil
                })
                if err != nil {
                    fmt.Printf("error walking the path %v: %v\n", filepath.Join(dir, d.Name()), err)
                }
            }(d)
        }
    }

    wg.Wait()
    close(fileChan)

    progressContainer.Hide()
}
