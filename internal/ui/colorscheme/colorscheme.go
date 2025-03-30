package colorscheme

import (
	"fmt"
	"os"
	"strings"

	"github.com/muesli/termenv"
)

// ColorFunc is a function that applies color to a string
type ColorFunc func(string) string

// Scheme defines a complete color scheme for the UI
type Scheme struct {
	Name        string
	Description string

	// UI Elements
	Header     ColorFunc
	Subheader  ColorFunc
	Command    ColorFunc
	Label      ColorFunc
	Value      ColorFunc
	Progress   ColorFunc
	Completed  ColorFunc
	Error      ColorFunc
	Comparison ColorFunc
	Faster     ColorFunc
	Slower     ColorFunc
}

// Get the terminal's color profile and determine if it supports RGB colors
var (
	output      = termenv.NewOutput(os.Stdout)
	profile     = output.ColorProfile()
	supportsRGB = profile == termenv.TrueColor
)

// Helper function to create a styled text function
func colorize(hex string, bold bool) ColorFunc {
	// For ASCII terminals, fall back to monochrome
	if profile == termenv.Ascii {
		if bold {
			return func(s string) string { return termenv.String(s).Bold().String() }
		}
		return func(s string) string { return s }
	}

	// For ANSI and ANSI256, we can still use colors but might need to approximate
	style := termenv.Style{}

	// Apply appropriate color based on terminal capabilities
	if profile == termenv.TrueColor {
		// Full RGB support
		style = style.Foreground(profile.Color(hex))
	} else if profile == termenv.ANSI256 {
		// 256 color support - termenv will approximate the color
		style = style.Foreground(profile.Color(hex))
	} else if profile == termenv.ANSI {
		// Basic 16 color support - termenv will approximate the color
		style = style.Foreground(profile.Color(hex))
	}

	if bold {
		style = style.Bold()
	}
	return style.Styled
}

// IsDarkBackground detects if the terminal has a dark background
func IsDarkBackground() bool {
	// Use termenv's built-in background detection
	bg := output.BackgroundColor()

	// Convert background color to RGB
	if c, ok := bg.(termenv.ANSI256Color); ok {
		// For ANSI 256 colors, we can check the color index
		// Dark backgrounds typically use colors 0, 16-27, 232-243
		idx := int(c)
		if idx == 0 || (idx >= 16 && idx <= 27) || (idx >= 232 && idx <= 243) {
			return true
		}
		return false
	}

	// For RGB colors, calculate perceived brightness
	// using the formula: (0.299*R + 0.587*G + 0.114*B)
	if c, ok := bg.(termenv.RGBColor); ok {
		// Get the hex color string and parse it
		hex := string(c)
		if len(hex) >= 7 {
			// Parse hex color #RRGGBB
			var r, g, b int
			fmt.Sscanf(hex[1:], "%02x%02x%02x", &r, &g, &b)
			brightness := (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 255.0
			return brightness < 0.5
		}
	}

	// Default to dark background if we can't determine
	return true
}

// GetAdaptiveScheme returns a color scheme that adapts to the terminal background
func GetAdaptiveScheme() *Scheme {
	if IsDarkBackground() {
		return Catppuccin() // Dark theme
	} else {
		return SolarizedLight() // Light theme
	}
}

// GetScheme returns a color scheme by name
func GetScheme(name string) (*Scheme, error) {
	name = strings.ToLower(name)

	// If terminal doesn't support RGB colors, fall back to monochrome
	// Only fall back if not explicitly requesting monochrome
	if !supportsRGB && name != "monochrome" {
		return Monochrome(), nil
	}

	switch name {
	case "default":
		return Default(), nil
	case "auto":
		return GetAdaptiveScheme(), nil
	case "catppuccin":
		return Catppuccin(), nil
	case "tokyonight":
		return TokyoNight(), nil
	case "nord":
		return Nord(), nil
	case "monokai":
		return Monokai(), nil
	case "solarized":
		return Solarized(), nil
	case "solarized-light":
		return SolarizedLight(), nil
	case "gruvbox":
		return Gruvbox(), nil
	case "monochrome":
		return Monochrome(), nil
	default:
		return nil, fmt.Errorf("unknown color scheme: %s", name)
	}
}

// ListSchemes returns a list of available color scheme names
func ListSchemes() []string {
	return []string{
		"default",
		"auto",
		"catppuccin",
		"tokyonight",
		"nord",
		"monokai",
		"solarized",
		"solarized-light",
		"gruvbox",
		"monochrome",
	}
}

// Default returns the default color scheme
func Default() *Scheme {
	return Catppuccin()
}

// FormatSchemeList returns a formatted string listing all available color schemes
func FormatSchemeList() string {
	var sb strings.Builder
	sb.WriteString("Available color schemes:\n")

	for _, name := range ListSchemes() {
		scheme, _ := GetScheme(name)
		sb.WriteString(fmt.Sprintf("  - %s: %s\n", name, scheme.Description))
	}

	return sb.String()
}
