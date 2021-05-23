package output

import (
	"bufio"
	"os"
	"strconv"

	"github.com/eur0pa/ffuf/pkg/ffuf"
)

func writeTXT(filename string, config *ffuf.Config, res []ffuf.Result) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	for _, r := range res {
		_, err := w.WriteString(toTXT(r))
		if err != nil {
			return err
		}
	}
	return nil
}

func toTXT(r ffuf.Result) string {
	res := strconv.FormatInt(r.StatusCode, 10) + " " +
		strconv.FormatInt(r.ContentLength, 10) + " " +
		strconv.FormatInt(r.ContentWords, 10) + " " +
		strconv.FormatInt(r.ContentLines, 10) + " " +
		r.Url

	if r.RedirectLocation != "" {
		res += " -> " + r.RedirectLocation
	}
	res += "\n"

	return res
}
