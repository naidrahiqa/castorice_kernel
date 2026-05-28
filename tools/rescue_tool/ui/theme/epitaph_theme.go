package theme

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type EpitaphTheme struct{}

func NewEpitaphTheme() fyne.Theme {
	return &EpitaphTheme{}
}

// Color returns the color for a specific resource name and theme variant
func (t *EpitaphTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	// Custom palette based on Epitaph Kernel branding (Premium Indigo/Dark Theme)
	switch name {
	case theme.ColorNameBackground:
		return color.RGBA{R: 0x0d, G: 0x11, B: 0x17, A: 0xff} // Deep dark background #0d1117
	case theme.ColorNameInputBackground, theme.ColorNameOverlayBackground:
		return color.RGBA{R: 0x16, G: 0x1b, B: 0x22, A: 0xff} // Surface #161b22
	case theme.ColorNamePrimary:
		return color.RGBA{R: 0x63, G: 0x66, B: 0xf1, A: 0xff} // Indigo accent #6366f1
	case theme.ColorNameFocus:
		return color.RGBA{R: 0x06, G: 0xb6, B: 0xd4, A: 0x7f} // Cyan focus glow #06b6d4 (half transparent)
	case theme.ColorNameSelection:
		return color.RGBA{R: 0x63, G: 0x66, B: 0xf1, A: 0x3f} // Selection background (light indigo)
	case theme.ColorNameForeground:
		return color.RGBA{R: 0xe2, G: 0xe8, B: 0xf0, A: 0xff} // Slate-200 primary text
	case theme.ColorNamePlaceHolder:
		return color.RGBA{R: 0x64, G: 0x74, B: 0x8b, A: 0xff} // Slate-500 muted text
	case theme.ColorNameSuccess:
		return color.RGBA{R: 0x22, G: 0xc5, B: 0x5e, A: 0xff} // Green
	case theme.ColorNameWarning:
		return color.RGBA{R: 0xf5, G: 0x9e, B: 0x0b, A: 0xff} // Amber
	case theme.ColorNameError:
		return color.RGBA{R: 0xef, G: 0x44, B: 0x44, A: 0xff} // Red
	case theme.ColorNameButton:
		return color.RGBA{R: 0x21, G: 0x26, B: 0x2d, A: 0xff} // Button surface
	case theme.ColorNameDisabled:
		return color.RGBA{R: 0x21, G: 0x26, B: 0x2d, A: 0x7f}
	case theme.ColorNameDisabledButton:
		return color.RGBA{R: 0x16, G: 0x1b, B: 0x22, A: 0xff}
	case theme.ColorNameSeparator:
		return color.RGBA{R: 0x30, G: 0x36, B: 0x3d, A: 0xff} // Border color
	}

	return theme.DefaultTheme().Color(name, theme.VariantDark)
}

func (t *EpitaphTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *EpitaphTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *EpitaphTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 14
	case theme.SizeNameHeadingText:
		return 18
	case theme.SizeNameSubHeadingText:
		return 16
	case theme.SizeNameCaptionText:
		return 11
	case theme.SizeNameInlineIcon:
		return 20
	case theme.SizeNamePadding:
		return 8
	case theme.SizeNameScrollBar:
		return 8
	case theme.SizeNameSelectionRadius:
		return 4
	case theme.SizeNameInputRadius:
		return 6
	}
	return theme.DefaultTheme().Size(name)
}

// Icon helper functions mapping to appropriate Fyne icons

func HomeIcon() fyne.Resource {
	return theme.HomeIcon()
}

func RescueIcon() fyne.Resource {
	return theme.WarningIcon()
}

func LogIcon() fyne.Resource {
	return theme.DocumentIcon()
}

func WifiIcon() fyne.Resource {
	// Use SettingsIcon or SearchIcon as a placeholder
	return theme.SettingsIcon()
}

func ValidateIcon() fyne.Resource {
	return theme.ConfirmIcon()
}

func SearchIcon() fyne.Resource {
	return theme.SearchIcon()
}

func SettingsIcon() fyne.Resource {
	return theme.SettingsIcon()
}

