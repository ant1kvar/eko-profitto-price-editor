package main

import (
	"fmt"
	"regexp"
	"strings"
)

func ExtractTableData(html string) ([]string, []string, [][]string) {
	tableStart := strings.Index(html, `<table class="table table-bordered" id="PriceTable">`)
	tableEnd := strings.Index(html, "</table>")
	if tableStart == -1 || tableEnd == -1 {
		return nil, nil, nil
	}
	tableHtml := html[tableStart : tableEnd+len("</table>")]

	rowRe := regexp.MustCompile(`(?s)<tr.*?>(.*?)</tr>`)
	cellRe := regexp.MustCompile(`<t[dh][^>]*>(.*?)</t[dh]>`)
	rows := rowRe.FindAllStringSubmatch(tableHtml, -1)

	var headers []string
	var periods []string
	var data [][]string

	for i, row := range rows {
		cells := cellRe.FindAllStringSubmatch(row[1], -1)
		var values []string
		for _, c := range cells {
			values = append(values, stripTags(c[1]))
		}
		if i == 0 {
			headers = values[1:]
			continue
		}
		periods = append(periods, values[0])
		data = append(data, values[1:])
	}

	return periods, headers, data
}

func stripTags(s string) string {
	tag := regexp.MustCompile(`<.*?>`)
	return tag.ReplaceAllString(s, "")
}

func UpdateTable(original string, periods []string, headers []string, newData [][]string) string {
	start := strings.Index(original, `<div class="table-responsive">`)
	if start == -1 {
		return original
	}

	mobileStart := strings.Index(original[start:], `<div class="mobile-cards">`)
	mobileEnd := strings.Index(original[start+mobileStart:], `</div>`)
	if mobileStart == -1 || mobileEnd == -1 {
		return original
	}
	end := start + mobileStart + mobileEnd + len("</div>")

	var desktopTable strings.Builder
	desktopTable.WriteString(`<div class="table-responsive">
<table class="table table-bordered" id="PriceTable">
<tr><td>Период</td>`)
	for _, h := range headers {
		desktopTable.WriteString("<td>" + h + "</td>")
	}
	desktopTable.WriteString("</tr>")
	for i, row := range newData {
		desktopTable.WriteString(`<tr class="price-tr"><td>` + periods[i] + `</td>`)
		for _, val := range row {
			desktopTable.WriteString("<td>" + val + "</td>")
		}
		desktopTable.WriteString("</tr>")
	}
	desktopTable.WriteString("</table></div>")

	var mobile strings.Builder
	mobile.WriteString(`<div class="mobile-cards">`)
	for i, row := range newData {
		mobile.WriteString(fmt.Sprintf(`
<table class="period-table">
<thead><tr><th colspan="2">%s</th></tr></thead>
<tbody>
`, periods[i]))
		for j, val := range row {
			price := strings.TrimSpace(val)
			if price == "" {
				price = "-"
			}
			mobile.WriteString(fmt.Sprintf("<tr><td>%s</td><td>%s ₽</td></tr>\n", headers[j], price))
		}
		mobile.WriteString("</tbody></table>")
	}
	mobile.WriteString("\n</div>")

	return original[:start] + desktopTable.String() + "\n" + mobile.String() + original[end:]
}

func ExtractNotes(html string) string {
	re := regexp.MustCompile(`(?s)<p><b>.*?</p>`)
	match := re.FindString(html)
	return stripTags(match)
}

func UpdateNotes(original, newNote string) string {
	re := regexp.MustCompile(`(?s)<p><b>.*?</p>`)
	newHtml := `<p><b>` + strings.ReplaceAll(newNote, "\n", "<br>") + `</b></p>`
	return re.ReplaceAllString(original, newHtml)
}
