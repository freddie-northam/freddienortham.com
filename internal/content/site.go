package content

import "time"

// Config holds the handful of facts that appear everywhere. Values marked
// TODO are placeholders: the build prints a warning for each one still unset
// (see build.WarnPlaceholders) so none of them can reach production unnoticed.
type Config struct {
	Name        string
	Role        string
	Headline    string
	Status      string
	Domain      string
	BaseURL     string
	Description string
	Email       string
	GitHub      string
	X           string
	LinkedIn    string
	BirthDate   *time.Time
	NoIndex     bool
	Colophon    string
}

// Placeholder marks a value Freddie still needs to supply. Never invent these.
const Placeholder = "TODO"

var birth = mustDate("2003-01-01") // TODO: real birth date

var Site = Config{
	Name: "Freddie Northam",
	Role: "Co-founder & Head of Product at XALT",

	// The homepage opens with a sentence, not a name. A name at 96px says
	// nothing; this says what he does and the one fact nobody expects.
	Headline: "Co-founder and Head of Product at XALT, building the tools brands use to reach their fans. I started at 14 and never went to university.",
	Status:   "Building at XALT",

	Domain:      "freddienortham.com",
	BaseURL:     "https://freddienortham.com",
	Description: "Co-founder and Head of Product at XALT. Writing, projects, and everything else.",

	Email:    Placeholder, // TODO: real address for the footer mailto
	GitHub:   "https://github.com/freddie-northam",
	X:        Placeholder, // TODO: real X profile URL
	LinkedIn: "https://www.linkedin.com/in/frederick-northam/",

	BirthDate: &birth, // TODO: replace with the real date

	// noindex stays on until explicitly flipped, so Google never gets hold of
	// placeholder copy and keeps it for months.
	NoIndex: true,

	// The July 2026 Jacobian conjecture counterexample: a map whose Jacobian
	// determinant is a non-zero constant everywhere, and which is still not
	// invertible. Unattributed and unexplained on purpose.
	Colophon: "locally everywhere does not imply everywhere",
}

func mustDate(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}

// Placeholders returns the config fields still awaiting real values.
func (c Config) Placeholders() []string {
	var out []string
	for name, v := range map[string]string{"Email": c.Email, "X": c.X} {
		if v == Placeholder {
			out = append(out, name)
		}
	}
	return out
}
