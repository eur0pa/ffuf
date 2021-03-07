package input

import (
	"bufio"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/eur0pa/ffuf/pkg/ffuf"
	tld "github.com/jpillora/go-tld"
)

type WordlistInput struct {
	config    *ffuf.Config
	data      [][]byte
	position  int
	keyword   string
	templates map[string]string
}

func NewWordlistInput(keyword string, value string, conf *ffuf.Config) (*WordlistInput, error) {
	var wl WordlistInput
	wl.keyword = keyword
	wl.config = conf
	wl.position = 0
	wl.templates = make(map[string]string)
	var valid bool
	var err error
	// templated replacements
	t := time.Now()
	wl.templates["YYYY"] = t.Format("2006")
	wl.templates["YY"] = t.Format("06")
	wl.templates["MM"] = t.Format("01")
	wl.templates["DD"] = t.Format("02")

	u, err := tld.Parse(conf.Url)
	if err == nil {
		wl.templates["SUB"] = u.Subdomain
		wl.templates["HOST"] = u.Domain
		wl.templates["TLD"] = u.TLD
	}

	// stdin?
	if value == "-" {
		// yes
		valid = true
	} else {
		// no
		valid, err = wl.validFile(value)
	}
	if err != nil {
		return &wl, err
	}
	if valid {
		err = wl.readFile(value)
	}

	return &wl, err
}

//Position will return the current position in the input list
func (w *WordlistInput) Position() int {
	return w.position
}

//ResetPosition resets the position back to beginning of the wordlist.
func (w *WordlistInput) ResetPosition() {
	w.position = 0
}

//Keyword returns the keyword assigned to this InternalInputProvider
func (w *WordlistInput) Keyword() string {
	return w.keyword
}

//Next will increment the cursor position, and return a boolean telling if there's words left in the list
func (w *WordlistInput) Next() bool {
	return w.position < len(w.data)
}

//IncrementPosition will increment the current position in the inputprovider data slice
func (w *WordlistInput) IncrementPosition() {
	w.position += 1
}

//Value returns the value from wordlist at current cursor position
func (w *WordlistInput) Value() []byte {
	return w.data[w.position]
}

//Total returns the size of wordlist
func (w *WordlistInput) Total() int {
	return len(w.data)
}

//validFile checks that the wordlist file exists and can be read
func (w *WordlistInput) validFile(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	f.Close()
	return true, nil
}

//readFile reads the file line by line to a byte slice
func (w *WordlistInput) readFile(path string) error {
	var file *os.File
	var err error
	if path == "-" {
		file = os.Stdin
	} else {
		file, err = os.Open(path)
		if err != nil {
			return err
		}
	}
	defer file.Close()

	var data [][]byte
	var data2 [][]byte
	var ok bool
	uniq := make(map[string]struct{})
	reader := bufio.NewScanner(file)
	re := regexp.MustCompile(`%EXT%`)
	re2 := regexp.MustCompile(`%EXT2%`)
	for reader.Scan() {
		text := replaceTemplates(reader.Text(), w.templates)
		textB := []byte(text)
		if w.config.DirSearchCompat && len(w.config.Extensions) > 0 {
			if len(w.config.Extensions2) > 0 {
				if re2.Match(textB) {
					for _, ext := range w.config.Extensions2 {
						contnt := re2.ReplaceAll(textB, []byte(ext))
						data = append(data, []byte(contnt))
					}
				}
			}
			if re.Match(textB) {
				for _, ext := range w.config.Extensions {
					contnt := re.ReplaceAll(textB, []byte(ext))
					data = append(data, []byte(contnt))
				}
			} else {
				if w.config.IgnoreWordlistComments {
					text, ok = stripComments(text)
					if !ok {
						continue
					}
				}
				data = append(data, []byte(text))
			}
		} else {
			if w.config.IgnoreWordlistComments {
				text, ok = stripComments(text)
				if !ok {
					continue
				}
			}
			data = append(data, []byte(text))
			if w.keyword == "FUZZ" && len(w.config.Extensions) > 0 {
				for _, ext := range w.config.Extensions {
					data = append(data, []byte(text+ext))
				}
			}
		}
	}
	// Deduplicate
	for _, line := range data {
		x := string(line)
		if _, ok := uniq[x]; !ok {
			uniq[x] = struct{}{}
		}
	}
	for line := range uniq {
		data2 = append(data2, []byte(line))
	}
	w.data = data2
	return reader.Err()
}

// stripComments removes all kind of comments from the word
func stripComments(text string) (string, bool) {
	// If the line starts with a # ignoring any space on the left,
	// return blank.
	if strings.HasPrefix(strings.TrimLeft(text, " "), "#") {
		return "", false
	}

	// If the line has # later after a space, that's a comment.
	// Only send the word upto space to the routine.
	index := strings.Index(text, " #")
	if index == -1 {
		return text, true
	}
	return text[:index], true
}

// replaceTemplates performs the templated dynamic replacements
func replaceTemplates(text string, t map[string]string) string {
	if strings.Contains(text, "{YYYY}") {
		if t["YYYY"] != "" {
			text = strings.ReplaceAll(text, "{YYYY}", t["YYYY"])
		} else {
			return ""
		}
	}
	if strings.Contains(text, "{YY}") {
		if t["YY"] != "" {
			text = strings.ReplaceAll(text, "{YY}", t["YY"])
		} else {
			return ""
		}
	}
	if strings.Contains(text, "{MM}") {
		if t["MM"] != "" {
			text = strings.ReplaceAll(text, "{MM}", t["MM"])
		} else {
			return ""
		}
	}
	if strings.Contains(text, "{DD}") {
		if t["DD"] != "" {
			text = strings.ReplaceAll(text, "{DD}", t["DD"])
		} else {
			return ""
		}
	}
	if strings.Contains(text, "{SUB}") {
		if t["SUB"] != "" {
			text = strings.ReplaceAll(text, "{SUB}", t["SUB"])
		} else {
			return ""
		}
	}
	if strings.Contains(text, "{HOST}") {
		if t["HOST"] != "" {
			text = strings.ReplaceAll(text, "{HOST}", t["HOST"])
		} else {
			return ""
		}
	}
	if strings.Contains(text, "{TLD}") {
		if t["TLD"] != "" {
			text = strings.ReplaceAll(text, "{TLD}", t["TLD"])
		} else {
			return ""
		}
	}

	return text
}
