package pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// NewPlaceholderPage creates a clean placeholder view for coming-soon tabs
func NewPlaceholderPage(title, subtitle string) fyne.CanvasObject {
	lblTitle := widget.NewLabel(title)
	lblTitle.TextStyle = fyne.TextStyle{Bold: true}
	
	lblSub := widget.NewLabel(subtitle)
	lblSub.Alignment = fyne.TextAlignCenter
	
	box := container.NewCenter(
		container.NewVBox(
			lblTitle,
			lblSub,
		),
	)
	
	return box
}
