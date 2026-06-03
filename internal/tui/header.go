package tui

import forgeui "github.com/adaouat/forge/ui"

const asciiArt = `
██████╗ ██╗███████╗██████╗  ██████╗ ███████╗████████╗
██╔══██╗██║██╔════╝██╔══██╗██╔═══██╗██╔════╝╚══██╔══╝
██████╔╝██║█████╗  ██████╔╝██║   ██║███████╗   ██║
██╔══██╗██║██╔══╝  ██╔══██╗██║   ██║╚════██║   ██║
██████╔╝██║██║     ██║  ██║╚██████╔╝███████║   ██║
╚═════╝ ╚═╝╚═╝     ╚═╝  ╚═╝ ╚═════╝ ╚══════╝   ╚═╝`

// CatchPhrase is the tagline displayed under the logo. TBD.
const CatchPhrase = "<catch phrase>"

// HelpLong returns the ASCII art + tagline for use as cobra root.Long.
// Rendered by fang with 2-space left padding as the help description.
func HelpLong() string { return forgeui.HelpLong(asciiArt, CatchPhrase) }

// VersionTemplate returns a cobra text/template string for --version output.
// cobra fills {{.Name}} and {{.Version}} at runtime.
func VersionTemplate() string { return forgeui.VersionTemplate(asciiArt, CatchPhrase) }
