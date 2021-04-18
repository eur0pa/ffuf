package output

import (
	"html/template"
	"os"
	"time"

	"github.com/eur0pa/ffuf/pkg/ffuf"
)

type htmlFileOutput struct {
	CommandLine string
	Time        string
	Keys        []string
	Results     []ffuf.Result
}

const (
	htmlTemplate = `
<!DOCTYPE html>
<html>
  <head>
    <meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
    <meta
      name="viewport"
      content="width=device-width, initial-scale=1, maximum-scale=1.0"
    />
    <title>FFUF Report - </title>

    <!-- CSS  -->
    <link
      href="https://fonts.googleapis.com/icon?family=Material+Icons"
      rel="stylesheet"
    />
    <link
      rel="stylesheet"
      href="https://cdnjs.cloudflare.com/ajax/libs/materialize/1.0.0/css/materialize.min.css"
	/>
	<link 
	  rel="stylesheet" 
	  type="text/css" 
	  href="https://cdn.datatables.net/1.10.20/css/jquery.dataTables.css"
	/>
  
  </head>

  <body>
    <main class="section no-pad-bot" id="index-banner">
      <div class="container">

      <table id="ffufreport">
        <thead>
          <tr>
              <th>Status</th>
              <th>Length</th>
              <th>Words</th>
			  <th>Lines</th>
			  <th>URL</th>
			  <th>Redirect location</th>
          </tr>
        </thead>

        <tbody>
			{{range $result := .Results}}
                <tr class="result-{{ $result.StatusCode }}" style="background-color: {{$result.HTMLColor}};">
                    <td><font color="black" class="status-code">{{ $result.StatusCode }}</font></td>
                    <td>{{ $result.ContentLength }}</td>
                    <td>{{ $result.ContentWords }}</td>
					<td>{{ $result.ContentLines }}</td>
                    <td><a href="{{ $result.Url }}">{{ $result.Url }}</a></td>
                    <td><a href="{{ $result.RedirectLocation }}">{{ $result.RedirectLocation }}</a></td>
                </tr>
            {{ end }}
        </tbody>
      </table>

        </div>
      </div>
    </main>

    <!--JavaScript at end of body for optimized loading-->
	<script src="https://code.jquery.com/jquery-3.4.1.min.js" integrity="sha256-CSXorXvZcTkaix6Yvo6HppcZGetbYMGWSFlBw8HfCJo=" crossorigin="anonymous"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/materialize/1.0.0/js/materialize.min.js"></script>
    <script type="text/javascript" charset="utf8" src="https://cdn.datatables.net/1.10.20/js/jquery.dataTables.js"></script>
    <script>
    $(document).ready(function() {
        $('#ffufreport').DataTable(
            {
                "aLengthMenu": [
                    [250, 500, 1000, 2500, -1],
                    [250, 500, 1000, 2500, "All"]
                ]
            }
        )
        $('select').formSelect();
        });
    </script>
    <style>
      body {
        display: flex;
        min-height: 100vh;
        flex-direction: column;
      }

      main {
        flex: 1 0 auto;
      }
    </style>
  </body>
</html>

	`
)

// colorizeResults returns a new slice with HTMLColor attribute
func colorizeResults(results []ffuf.Result) []ffuf.Result {
	newResults := make([]ffuf.Result, 0)

	for _, r := range results {
		result := r
		result.HTMLColor = "black"

		s := result.StatusCode

		if s >= 200 && s <= 299 {
			result.HTMLColor = "#adea9e"
		}

		if s >= 300 && s <= 399 {
			result.HTMLColor = "#bbbbe6"
		}

		if s >= 400 && s <= 499 {
			result.HTMLColor = "#d2cb7e"
		}

		if s >= 500 && s <= 599 {
			result.HTMLColor = "#de8dc1"
		}

		newResults = append(newResults, result)
	}

	return newResults
}

func writeHTML(filename string, config *ffuf.Config, results []ffuf.Result) error {

	if config.OutputCreateEmptyFile && (len(results) == 0) {
		return nil
	}

	results = colorizeResults(results)

	ti := time.Now()

	keywords := make([]string, 0)
	for _, inputprovider := range config.InputProviders {
		keywords = append(keywords, inputprovider.Keyword)
	}

	outHTML := htmlFileOutput{
		CommandLine: config.CommandLine,
		Time:        ti.Format(time.RFC3339),
		Results:     results,
		Keys:        keywords,
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	templateName := "output.html"
	t := template.New(templateName).Delims("{{", "}}")
	_, err = t.Parse(htmlTemplate)
	if err != nil {
		return err
	}
	err = t.Execute(f, outHTML)
	return err
}
