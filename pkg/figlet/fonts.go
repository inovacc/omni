package figlet

import "embed"

//go:embed fonts/*.flf
var embeddedFonts embed.FS

// LoadEmbedded loads a named font from the embedded font collection.
// The name should not include the .flf extension.
func LoadEmbedded(name string) (*Font, error) {
	data, err := embeddedFonts.ReadFile("fonts/" + name + ".flf")
	if err != nil {
		return nil, err
	}

	return LoadFont(data)
}

// ListFonts returns the names of all embedded fonts.
func ListFonts() []string {
	entries, err := embeddedFonts.ReadDir("fonts")
	if err != nil {
		return nil
	}

	var names []string

	for _, e := range entries {
		name := e.Name()
		if len(name) > 4 && name[len(name)-4:] == ".flf" {
			names = append(names, name[:len(name)-4])
		}
	}

	return names
}
