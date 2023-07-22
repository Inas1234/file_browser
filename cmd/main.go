package main

import (
	"io/ioutil"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func listFiles(dir string) ([]os.FileInfo, error) {
	return ioutil.ReadDir(dir)
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("File Browser")

	myWindow.Resize(fyne.NewSize(800, 600))

	input := widget.NewEntry()
	input.SetPlaceHolder("Enter a URL")
	myWindow.SetContent(container.NewVBox(input, widget.NewButton("OK", func() {
		list, err := listFiles(input.Text)
		if err != nil {
			panic(err)
		}
		var items []fyne.CanvasObject
		for _, item := range list {
			name := item.Name()
			button := widget.NewButton(name, func() {
				if item.IsDir() {
					list, err = listFiles(name)
					if err != nil {
						panic(err)
					}
					items = nil
					for _, item := range list {
						button := widget.NewButton(item.Name(), nil)
						items = append(items, button)
					}
					myWindow.SetContent(container.NewVBox(items...))
				}
			})
			items = append(items, button)
		}

		content := container.NewVBox(items...)
		myWindow.SetContent(content)
	})))
	
	myWindow.ShowAndRun()
}
