package pages

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

var (
	colorBackground = color.RGBA{R: 0x07, G: 0x09, B: 0x0e, A: 0xff}
	colorSurface    = color.RGBA{R: 0x12, G: 0x16, B: 0x20, A: 0xff}
	colorPrimary    = color.RGBA{R: 0x63, G: 0x66, B: 0xf1, A: 0xff}
	colorSecondary  = color.RGBA{R: 0x06, G: 0xb6, B: 0xd4, A: 0xff}
	colorSuccess    = color.RGBA{R: 0x10, G: 0xb9, B: 0x81, A: 0xff}
	colorWarning    = color.RGBA{R: 0xf5, G: 0x9e, B: 0x0b, A: 0xff}
	colorError      = color.RGBA{R: 0xef, G: 0x44, B: 0x44, A: 0xff}
	colorTextPri    = color.RGBA{R: 0xf1, G: 0xf5, B: 0xf9, A: 0xff}
	colorTextMuted  = color.RGBA{R: 0x94, G: 0xa3, B: 0xb8, A: 0xff}
	colorBorder     = color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xff}
)

type neoLayout struct {
	offset float32
}

func (n *neoLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 2 {
		return
	}
	objects[0].Resize(fyne.NewSize(size.Width-n.offset, size.Height-n.offset))
	objects[0].Move(fyne.NewPos(n.offset, n.offset))
	objects[1].Resize(fyne.NewSize(size.Width-n.offset, size.Height-n.offset))
	objects[1].Move(fyne.NewPos(0, 0))
}

func (n *neoLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) < 2 {
		return fyne.NewSize(150, 100)
	}
	min := objects[1].MinSize()
	return fyne.NewSize(min.Width+n.offset, min.Height+n.offset)
}

func NewNeoCard(title string, subtitle string, content fyne.CanvasObject) *fyne.Container {
	shadow := canvas.NewRectangle(color.Black)

	titleLbl := widget.NewLabel(title)
	titleLbl.TextStyle = fyne.TextStyle{Bold: true}

	headerBox := container.NewVBox(
		container.NewHBox(titleLbl, layout.NewSpacer()),
	)
	if subtitle != "" {
		subLbl := widget.NewLabel(subtitle)
		subLbl.TextStyle = fyne.TextStyle{Italic: true}
		headerBox.Add(subLbl)
	}

	contentPadded := container.NewPadded(content)

	var bodyContent fyne.CanvasObject
	if title != "" || subtitle != "" {
		sep := canvas.NewRectangle(color.Black)
		sep.SetMinSize(fyne.NewSize(0, 2))
		bodyContent = container.NewBorder(
			container.NewVBox(headerBox, sep), nil, nil, nil,
			contentPadded,
		)
	} else {
		bodyContent = contentPadded
	}

	bodyBg := canvas.NewRectangle(colorSurface)
	bodyBorder := canvas.NewRectangle(color.Black)

	cardBodyStack := container.New(layout.NewMaxLayout(),
		bodyBorder,
		container.NewPadded(container.New(layout.NewMaxLayout(), bodyBg, bodyContent)),
	)

	return container.New(&neoLayout{offset: 6}, shadow, cardBodyStack)
}

func NeoDivider() fyne.CanvasObject {
	r := canvas.NewRectangle(colorBorder)
	r.SetMinSize(fyne.NewSize(0, 1))
	return r
}

func NewNeoButton(text string, icon fyne.Resource, importance widget.ButtonImportance, onTap func()) *widget.Button {
	btn := widget.NewButtonWithIcon(text, icon, onTap)
	btn.Importance = importance
	return btn
}

func NewNeoHeading(text string) *widget.Label {
	lbl := widget.NewLabel(text)
	lbl.TextStyle = fyne.TextStyle{Bold: true}
	return lbl
}

func NewNeoLabel(text string) *widget.Label {
	lbl := widget.NewLabel(text)
	lbl.Wrapping = fyne.TextWrapWord
	return lbl
}
