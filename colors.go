package sawmill

import (
	"fmt"
	"strings"
)

// ColorScheme defines colors for different types and keys
type ColorScheme struct {
	Keys         string            // Default color for keys
	StringValues string            // Default color for string values
	IntValues    string            // Default color for integer values
	FloatValues  string            // Default color for float values
	BoolValues   string            // Default color for boolean values
	NullValues   string            // Default color for null values
	KeyMappings  map[string]string // Custom colors for specific keys (supports dot notation)
	Enabled      bool              // Whether coloring is enabled
}

// ANSI color codes
const (
	ColorReset   = "\033[0m"
	ColorRed     = "\033[31m"
	ColorGreen   = "\033[32m"
	ColorYellow  = "\033[33m"
	ColorBlue    = "\033[34m"
	ColorMagenta = "\033[35m"
	ColorCyan    = "\033[36m"
	ColorWhite   = "\033[37m"
	ColorBold    = "\033[1m"

	// Bright colors
	ColorBrightRed     = "\033[91m"
	ColorBrightGreen   = "\033[92m"
	ColorBrightYellow  = "\033[93m"
	ColorBrightBlue    = "\033[94m"
	ColorBrightMagenta = "\033[95m"
	ColorBrightCyan    = "\033[96m"
	ColorBrightWhite   = "\033[97m"
)

// DefaultColorScheme returns the default color scheme
func DefaultColorScheme() *ColorScheme {
	return &ColorScheme{
		Keys:         ColorBlue,
		StringValues: ColorRed,
		IntValues:    ColorGreen,
		FloatValues:  ColorYellow,
		BoolValues:   ColorMagenta,
		NullValues:   ColorCyan,
		KeyMappings:  make(map[string]string),
		Enabled:      true,
	}
}

// NewColorScheme creates a new color scheme with custom mappings
func NewColorScheme(keyMappings map[string]string) *ColorScheme {
	scheme := DefaultColorScheme()
	if keyMappings != nil {
		scheme.KeyMappings = keyMappings
	}
	return scheme
}

// colorizeValue applies color to a value based on its type
func (cs *ColorScheme) colorizeValue(value interface{}) string {
	if !cs.Enabled {
		return fmt.Sprintf("%v", value)
	}

	if value == nil {
		return cs.NullValues + "null" + ColorReset
	}

	switch v := value.(type) {
	case string:
		return cs.StringValues + fmt.Sprintf("%q", v) + ColorReset
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return cs.IntValues + fmt.Sprintf("%v", v) + ColorReset
	case float32, float64:
		return cs.FloatValues + fmt.Sprintf("%v", v) + ColorReset
	case bool:
		return cs.BoolValues + fmt.Sprintf("%v", v) + ColorReset
	default:
		// For complex types, convert to string
		return cs.StringValues + fmt.Sprintf("%q", fmt.Sprintf("%v", v)) + ColorReset
	}
}

// colorizeKey applies color to a key, checking custom mappings first
func (cs *ColorScheme) colorizeKey(keyPath string) string {
	if !cs.Enabled {
		return keyPath
	}

	// Check for exact match in custom mappings
	if color, exists := cs.KeyMappings[keyPath]; exists {
		return color + keyPath + ColorReset
	}

	// Check for partial matches (e.g., "user" matches "user.profile.name")
	for mappedKey, color := range cs.KeyMappings {
		if strings.HasPrefix(keyPath, mappedKey+".") || strings.HasSuffix(keyPath, "."+mappedKey) {
			return color + keyPath + ColorReset
		}
	}

	// Use default key color
	return cs.Keys + keyPath + ColorReset
}

// colorizeJSON applies syntax highlighting to JSON output
func (cs *ColorScheme) colorizeJSON(jsonStr string) string {
	if !cs.Enabled {
		return jsonStr
	}

	result := []rune(jsonStr)
	var output strings.Builder

	i := 0
	for i < len(result) {
		char := result[i]

		switch char {
		case '"':
			// Handle quoted strings (could be keys or values)
			output.WriteRune('"') // Write opening quote
			i++                   // Skip opening quote

			// Find the closing quote and collect content
			contentStart := i
			for i < len(result) && result[i] != '"' {
				if result[i] == '\\' && i+1 < len(result) {
					i++ // Skip escaped character
				}
				i++
			}

			if i < len(result) {
				quotedContent := string(result[contentStart:i])

				// Check if this is a key (followed by colon after closing quote)
				isKey := false
				j := i + 1 // Start after the closing quote
				for j < len(result) && (result[j] == ' ' || result[j] == '\t' || result[j] == '\n') {
					j++
				}
				if j < len(result) && result[j] == ':' {
					isKey = true
				}

				if isKey {
					coloredKey := cs.colorizeKey(quotedContent)
					output.WriteString(coloredKey)
				} else {
					coloredValue := cs.StringValues + quotedContent + ColorReset
					output.WriteString(coloredValue)
				}

				output.WriteRune('"') // Write closing quote
				i++                   // Skip closing quote
			}
			i-- // Adjust for loop increment

		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-':
			// Handle numbers
			start := i
			for i < len(result) && (isDigit(result[i]) || result[i] == '.' || result[i] == '-' || result[i] == 'e' || result[i] == 'E' || result[i] == '+') {
				i++
			}
			number := string(result[start:i])
			if strings.Contains(number, ".") {
				coloredNumber := cs.FloatValues + number + ColorReset
				output.WriteString(coloredNumber)
			} else {
				coloredNumber := cs.IntValues + number + ColorReset
				output.WriteString(coloredNumber)
			}
			i-- // Back up one since the loop will increment

		case 't', 'f':
			// Handle boolean values
			if i+4 <= len(result) && string(result[i:i+4]) == "true" {
				coloredBool := cs.BoolValues + "true" + ColorReset
				output.WriteString(coloredBool)
				i += 3 // Skip "rue" (loop will increment for 't')
			} else if i+5 <= len(result) && string(result[i:i+5]) == "false" {
				coloredBool := cs.BoolValues + "false" + ColorReset
				output.WriteString(coloredBool)
				i += 4 // Skip "alse" (loop will increment for 'f')
			} else {
				output.WriteRune(char)
			}

		case 'n':
			// Handle null values
			if i+4 <= len(result) && string(result[i:i+4]) == "null" {
				coloredNull := cs.NullValues + "null" + ColorReset
				output.WriteString(coloredNull)
				i += 3 // Skip "ull" (loop will increment for 'n')
			} else {
				output.WriteRune(char)
			}

		default:
			// Preserve all other characters (including colons, commas, braces, etc.)
			output.WriteRune(char)
		}

		i++
	}

	return output.String()
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func replaceWithFunc(input, pattern string, replacer func(match, capture string) string) string {
	return input
}

// ColorizeAttributes formats attributes with color highlighting
func (cs *ColorScheme) ColorizeAttributes(attrs *RecursiveMap, format string) string {
	if !cs.Enabled {
		return attrs.String()
	}

	switch format {
	case "json":
		return cs.colorizeAttributesJSON(attrs, 0)
	case "flat":
		return cs.colorizeAttributesFlat(attrs)
	default:
		return cs.colorizeAttributesNested(attrs, 0)
	}
}

// colorizeAttributesJSON formats attributes as colored JSON
func (cs *ColorScheme) colorizeAttributesJSON(attrs *RecursiveMap, indent int) string {
	if attrs.IsEmpty() {
		return "{}"
	}

	var result strings.Builder
	indentStr := strings.Repeat("  ", indent)

	result.WriteString("{\n")

	keys := attrs.Keys()
	for i, key := range keys {
		result.WriteString(indentStr + "  ")

		// Colorize key
		coloredKey := cs.colorizeKey(key)
		result.WriteString(`"` + coloredKey + `": `)

		child := attrs.children[key]
		if child.IsLeaf() {
			// Colorize value
			coloredValue := cs.colorizeValue(child.value)
			result.WriteString(coloredValue)
		} else {
			// Recursive call for nested objects
			nestedJSON := cs.colorizeAttributesJSON(child, indent+1)
			result.WriteString(nestedJSON)
		}

		if i < len(keys)-1 {
			result.WriteString(",")
		}
		result.WriteString("\n")
	}

	result.WriteString(indentStr + "}")
	return result.String()
}

// colorizeAttributesFlat formats attributes in flat key=value format with colors
func (cs *ColorScheme) colorizeAttributesFlat(attrs *RecursiveMap) string {
	var parts []string

	attrs.Walk(func(path []string, value interface{}) {
		keyPath := strings.Join(path, ".")
		coloredKey := cs.colorizeKey(keyPath)
		coloredValue := cs.colorizeValue(value)
		parts = append(parts, coloredKey+"="+coloredValue)
	})

	return strings.Join(parts, " ")
}

// colorizeAttributesNested formats attributes in nested format with colors
func (cs *ColorScheme) colorizeAttributesNested(attrs *RecursiveMap, indent int) string {
	if attrs.IsEmpty() {
		return ""
	}

	var result strings.Builder
	indentStr := strings.Repeat("  ", indent)

	for key, child := range attrs.children {
		result.WriteString("\n" + indentStr)
		coloredKey := cs.colorizeKey(key)
		result.WriteString(coloredKey + ":")

		if child.IsLeaf() {
			coloredValue := cs.colorizeValue(child.value)
			result.WriteString(" " + coloredValue)
		} else {
			nestedResult := cs.colorizeAttributesNested(child, indent+1)
			result.WriteString(nestedResult)
		}
	}

	return result.String()
}

// ParseColorCode converts common color names to ANSI codes
func ParseColorCode(colorName string) string {
	switch strings.ToLower(colorName) {
	case "red":
		return ColorRed
	case "green":
		return ColorGreen
	case "yellow":
		return ColorYellow
	case "blue":
		return ColorBlue
	case "magenta":
		return ColorMagenta
	case "cyan":
		return ColorCyan
	case "white":
		return ColorWhite
	case "bright_red":
		return ColorBrightRed
	case "bright_green":
		return ColorBrightGreen
	case "bright_yellow":
		return ColorBrightYellow
	case "bright_blue":
		return ColorBrightBlue
	case "bright_magenta":
		return ColorBrightMagenta
	case "bright_cyan":
		return ColorBrightCyan
	case "bright_white":
		return ColorBrightWhite
	case "bold":
		return ColorBold
	default:
		// Allow direct ANSI codes
		return colorName
	}
}
