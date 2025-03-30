package colorscheme

import (
	"github.com/muesli/termenv"
)

// Catppuccin returns the Catppuccin color scheme (Mocha variant)
func Catppuccin() *Scheme {
	return &Scheme{
		Name:        "Catppuccin",
		Description: "Soothing pastel theme (Mocha variant)",
		Header:      colorize("#cba6f7", true),  // Mauve (bold)
		Subheader:   colorize("#b4befe", false), // Lavender
		Command:     colorize("#b4befe", true),  // Lavender (bold)
		Label:       colorize("#74c7ec", false), // Sapphire
		Value:       colorize("#cdd6f4", false), // Text
		Progress:    colorize("#94e2d5", false), // Teal
		Completed:   colorize("#a6e3a1", false), // Green
		Error:       colorize("#f38ba8", false), // Red
		Comparison:  colorize("#cba6f7", false), // Mauve
		Faster:      colorize("#a6e3a1", false), // Green
		Slower:      colorize("#eba0ac", false), // Maroon
	}
}

// SolarizedLight returns the Solarized Light color scheme
func SolarizedLight() *Scheme {
	return &Scheme{
		Name:        "Solarized Light",
		Description: "Precision colors for machines and people (light variant)",
		Header:      colorize("#268bd2", true),  // Blue (bold)
		Subheader:   colorize("#2aa198", false), // Cyan
		Command:     colorize("#859900", true),  // Green (bold)
		Label:       colorize("#6c71c4", false), // Violet
		Value:       colorize("#657b83", false), // Base00 (body text)
		Progress:    colorize("#b58900", false), // Yellow
		Completed:   colorize("#859900", false), // Green
		Error:       colorize("#dc322f", false), // Red
		Comparison:  colorize("#d33682", false), // Magenta
		Faster:      colorize("#859900", false), // Green
		Slower:      colorize("#cb4b16", false), // Orange
	}
}

// TokyoNight returns the Tokyo Night color scheme
func TokyoNight() *Scheme {
	return &Scheme{
		Name:        "Tokyo Night",
		Description: "A dark and elegant theme",
		Header:      colorize("#7aa2f7", true),  // Blue (bold)
		Subheader:   colorize("#7dcfff", false), // Cyan
		Command:     colorize("#bb9af7", true),  // Purple (bold)
		Label:       colorize("#9d7cd8", false), // Purple
		Value:       colorize("#c0caf5", false), // Text
		Progress:    colorize("#e0af68", false), // Yellow
		Completed:   colorize("#9ece6a", false), // Green
		Error:       colorize("#f7768e", false), // Red
		Comparison:  colorize("#7aa2f7", false), // Blue
		Faster:      colorize("#9ece6a", false), // Green
		Slower:      colorize("#f7768e", false), // Red
	}
}

// Nord returns the Nord color scheme
func Nord() *Scheme {
	return &Scheme{
		Name:        "Nord",
		Description: "Arctic, north-bluish color palette",
		Header:      colorize("#88c0d0", true),  // Frost 2 (bold)
		Subheader:   colorize("#81a1c1", false), // Frost 3
		Command:     colorize("#5e81ac", true),  // Frost 4 (bold)
		Label:       colorize("#8fbcbb", false), // Frost 1
		Value:       colorize("#eceff4", false), // Snow Storm 3
		Progress:    colorize("#ebcb8b", false), // Aurora 3 (yellow)
		Completed:   colorize("#a3be8c", false), // Aurora 4 (green)
		Error:       colorize("#bf616a", false), // Aurora 1 (red)
		Comparison:  colorize("#b48ead", false), // Aurora 5 (purple)
		Faster:      colorize("#a3be8c", false), // Aurora 4 (green)
		Slower:      colorize("#d08770", false), // Aurora 2 (orange)
	}
}

// Monokai returns the Monokai color scheme
func Monokai() *Scheme {
	return &Scheme{
		Name:        "Monokai",
		Description: "Vibrant and colorful theme",
		Header:      colorize("#e6db74", true),  // Yellow (bold)
		Subheader:   colorize("#fd971f", false), // Orange
		Command:     colorize("#ae81ff", true),  // Purple (bold)
		Label:       colorize("#66d9ef", false), // Blue
		Value:       colorize("#f8f8f2", false), // Text
		Progress:    colorize("#a6e22e", false), // Green
		Completed:   colorize("#a6e22e", false), // Green
		Error:       colorize("#f92672", false), // Red
		Comparison:  colorize("#e6db74", false), // Yellow
		Faster:      colorize("#a6e22e", false), // Green
		Slower:      colorize("#f92672", false), // Red
	}
}

// Solarized returns the Solarized color scheme (dark variant)
func Solarized() *Scheme {
	return &Scheme{
		Name:        "Solarized",
		Description: "Precision colors for machines and people (dark variant)",
		Header:      colorize("#268bd2", true),  // Blue (bold)
		Subheader:   colorize("#2aa198", false), // Cyan
		Command:     colorize("#859900", true),  // Green (bold)
		Label:       colorize("#6c71c4", false), // Violet
		Value:       colorize("#839496", false), // Base0
		Progress:    colorize("#b58900", false), // Yellow
		Completed:   colorize("#859900", false), // Green
		Error:       colorize("#dc322f", false), // Red
		Comparison:  colorize("#d33682", false), // Magenta
		Faster:      colorize("#859900", false), // Green
		Slower:      colorize("#cb4b16", false), // Orange
	}
}

// Gruvbox returns the Gruvbox color scheme (dark variant)
func Gruvbox() *Scheme {
	return &Scheme{
		Name:        "Gruvbox",
		Description: "Retro groove color scheme (dark variant)",
		Header:      colorize("#fabd2f", true),  // Bright Yellow (bold)
		Subheader:   colorize("#fabd2f", false), // Bright Yellow
		Command:     colorize("#fb4934", true),  // Bright Red (bold)
		Label:       colorize("#83a598", false), // Bright Blue
		Value:       colorize("#ebdbb2", false), // Foreground
		Progress:    colorize("#8ec07c", false), // Bright Aqua
		Completed:   colorize("#b8bb26", false), // Bright Green
		Error:       colorize("#fb4934", false), // Bright Red
		Comparison:  colorize("#d3869b", false), // Bright Purple
		Faster:      colorize("#b8bb26", false), // Bright Green
		Slower:      colorize("#d65d0e", false), // Orange
	}
}

// Monochrome returns a monochrome color scheme (no colors, just plain text)
func Monochrome() *Scheme {
	// Use plain text without any color styling
	plain := func(s string) string { return s }
	bold := func(s string) string {
		return termenv.String(s).Bold().String()
	}

	return &Scheme{
		Name:        "Monochrome",
		Description: "Simple black and white theme (no colors)",
		Header:      bold,
		Subheader:   plain,
		Command:     bold,
		Label:       plain,
		Value:       plain,
		Progress:    plain,
		Completed:   bold,
		Error:       plain,
		Comparison:  bold,
		Faster:      bold,
		Slower:      plain,
	}
}
