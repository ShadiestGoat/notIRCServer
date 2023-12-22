package export

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/shadiestgoat/notIRCServer/db"
	"github.com/shadiestgoat/notIRCServer/users"
	"gopkg.in/yaml.v2"
)

var exporters = map[string](func(msgs []*db.Msg, w io.Writer) error){
	"log": func(msgs []*db.Msg, w io.Writer) error {
		for _, m := range msgs {
			authorStr := m.Author

			if m.To != "*" {
				authorStr = fmt.Sprintf("Whisper %v -> %v", m.Author, m.To)
			}

			// Wow look at that speed! wow so much saving over fprint!!! I'm so cool and hot
			_, err := w.Write([]byte(authorStr + ": " + m.Content + "\n"))
			if err != nil {
				return err
			}
		}

		return nil
	},
	"json": func(msgs []*db.Msg, w io.Writer) error {
		return json.NewEncoder(w).Encode(msgs)
	},
	"json-pretty": func(msgs []*db.Msg, w io.Writer) error {
		e := json.NewEncoder(w)
		e.SetIndent("", "\t")
		return e.Encode(msgs)
	},
	"yaml": func(msgs []*db.Msg, w io.Writer) error {
		e := yaml.NewEncoder(w)
		if err := e.Encode(msgs); err != nil {
			return err
		}
		return e.Close()
	},
	"yaml-sep": func(msgs []*db.Msg, w io.Writer) error {
		e := yaml.NewEncoder(w)

		for _, m := range msgs {
			if err := e.Encode(m); err != nil {
				return err
			}
		}

		return e.Close()
	},
}

// A command to write the output to.
func Export(format string, u *users.User, w io.Writer) error {
	enc := exporters[format]

	if enc == nil {
		available := []string{}

		for v := range exporters {
			available = append(available, v)
		}

		return fmt.Errorf("No exporter named '%v'. Available exporters: %v", format, strings.Join(available, ", "))
	}

	allMsgs := db.GetMessages()

	return enc(db.FilterForUser(allMsgs, u), w)
}
