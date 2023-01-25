package table

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"encoding/json"

	"github.com/mattn/go-runewidth"
)

const (
	// minSpace is the number of spaces between the end of the longest value and
	// the start of the next row
	minSpace = 3
)

type Table interface {
	Add(row ...string)
 	AddNestedTable(row ...interface{})
	Print()
	PrintJson()
	String() string
	SetOutputFormat(outputFormat string)
}

type PrintableTable struct {
	writer        io.Writer
	headers       []string
	headerPrinted bool
	maxSizes      []int
	rows          [][]string //each row is single line
	outputFormat  string // Text, JSON, future formats.
}

type KeyValueTable struct {
	*PrintableTable
}

func NewTable(w io.Writer, headers []string) Table {
	return &PrintableTable{
		writer:   w,
		headers:  headers,
		maxSizes: make([]int, len(headers)),
		outputFormat: "text",
	}
}

// A new method specially for setting output format at creation time.
func NewNestedTable(w io.Writer, headers []string, outputFormat string) Table {
	newTable := NewTable(w, headers)
	newTable.SetOutputFormat(outputFormat)
	return newTable
}

// Used to convert a Row that has a table in it to a string row.
// I couldn't jsut change the Add function to be Add(row ...interface{}) because it would break users adding rows like
// in the TestEllipsisTable() unit test
func (t *PrintableTable) AddNestedTable(row ...interface{}) {
	stringified := make([]string, len(row))
	for i := 0; i < len(row); i ++ {
		stringified[i] = fmt.Sprintf("%v", row[i])
	}
	t.Add(stringified...)
}

func (t *PrintableTable) Add(row ...string) {
	var maxLines int

	var columns [][]string
	for _, value := range row {
		var lines []string

		// Dont split out the lines by \n because JSON will use marshalIndent and it will make a mess of the output
		if t.outputFormat == "json" {
			lines = []string{fmt.Sprintf("%s", value)}
		} else {
			lines = strings.Split(fmt.Sprintf("%s", value), "\n")
		}

		if len(lines) > maxLines {
			maxLines = len(lines)
		}
		columns = append(columns, lines)
	}

	for i := 0; i < maxLines; i++ {
		var row []string
		for _, col := range columns {
			if i >= len(col) {
				row = append(row, "")
			} else {
				row = append(row, col[i])
			}
		}
		t.rows = append(t.rows, row)
	}

	// Incase we have more columns in a row than headers, need to update maxSizes
	if len(row) > len(t.maxSizes) {
		t.maxSizes = make([]int, len(row))
	}
}

func (t *PrintableTable) String() string {
	oldBuffer := t.writer
	newBuffer := bytes.Buffer{}
	t.writer = &newBuffer
	switch t.outputFormat {
	case "json":
		t.PrintJson()
	default:
		t.Print()
	}
	output := newBuffer.String()
	t.writer = oldBuffer
	return output
}

func (t *PrintableTable) Print() {
	for _, row := range append(t.rows, t.headers) {
		t.calculateMaxSize(row)
	}

	if t.headerPrinted == false {
		t.printHeader()
		t.headerPrinted = true
	}

	for _, line := range t.rows {
		t.printRow(line)
	}

	t.rows = [][]string{}
}

func (t *PrintableTable) calculateMaxSize(row []string) {
	for index, value := range row {
		cellLength := runewidth.StringWidth(Decolorize(value))
		if t.maxSizes[index] < cellLength {
			t.maxSizes[index] = cellLength
		}
	}
}

func (t *PrintableTable) printHeader() {
	output := ""
	for col, value := range t.headers {
		output = output + t.cellValue(col, HeaderColor(value))
	}
	fmt.Fprintln(t.writer, output)
}

func (t *PrintableTable) printRow(row []string) {
	output := ""
	for columnIndex, value := range row {
		if columnIndex == 0 {
			value = TableContentHeaderColor(value)
		}

		output = output + t.cellValue(columnIndex, value)
	}
	fmt.Fprintln(t.writer, output)
}

func (t *PrintableTable) cellValue(col int, value string) string {
	padding := ""
	if col < len(t.maxSizes)-1 {
		padding = strings.Repeat(" ", t.maxSizes[col]-runewidth.StringWidth(Decolorize(value))+minSpace)
	}
	return fmt.Sprintf("%s%s", value, padding)
}


// Prints out a nicely/human formatted Json string instead of a table structure
func (t *PrintableTable) PrintJson() {
	total_col := len(t.headers)
	// total_row := len(t.rows) - 1
	// A special type of table that should have a slightly different JSON output
	if total_col == 2 && (strings.ToLower(t.headers[0]) == "key" && strings.ToLower(t.headers[1]) == "value") {
		t.PrintKeyValueJson()
		return
	}
	
	jsonString, err := json.MarshalIndent(t, "", "\t")
	// jsonString, err := json.Marshal(t)
	if err != nil {
		// fmt.Fprintf(t.writer, "THERE WAS AN ERROR2\n %v\n%s", err.Error(), t.rows)
		// If there are errors, just dump the whole thing as a string
		fmt.Fprintf(t.writer, "%s", t.rows)	
	} else {
		fmt.Fprintf(t.writer, "%s", jsonString)	
	}
	
	// mimic behavior of Print()
	t.rows = [][]string{}
	
}

func (t *PrintableTable) PrintKeyValueJson() {
	tmpTable := &KeyValueTable{t}
	jsonString, err := json.MarshalIndent(tmpTable, "", "\t")
	// fmt.Printf("JSON:String %s\n", jsonString)
	if err != nil {
		// fmt.Fprintf(t.writer, "THERE WAS AN ERROR1\n %v\n%v", err.Error(), tmpTable)
		// If there are errors, just dump the whole thing as a string
		fmt.Fprintf(t.writer, "%s", t.rows)
	} else {
		fmt.Fprintf(t.writer, "%s", jsonString)	
	}
	
	// mimic behavior of Print()
	t.rows = [][]string{}
}

func (t *PrintableTable) SetOutputFormat(outFormat string) {
	switch strings.ToLower(outFormat) {
	case "json":
		t.outputFormat = "json"
	default:
		t.outputFormat =  "text"
	}
}

func (t *PrintableTable) MarshalJSON() ([]byte, error) {
	
	// Create a list of the rows
	var tableList []map[string]interface{}
	var colHeader string
	
	for _, row := range t.rows {
		// Create a map for to represent a row
		rowMap := make(map[string]interface{})
		for x, point := range row {
			// Some columns might not have a header, or an empty header. They need to have a name in JSON
			if x > len(t.headers) - 1 || t.headers[x] == "" {
				colHeader = fmt.Sprintf("column_%d", (x + 1))
			} else {
				colHeader = t.headers[x]
			}

			// js needs to be redefined each loop otherwise weird JSON.unmarshal errors happen when a row contains 0
			// see TestSLSubnetListProblem unit test
			var js json.RawMessage
			// Detects JSON styled strings, or numbers
			if json.Unmarshal([]byte(point), &js) == nil {
				rowMap[colHeader] = js	
			} else {
				rowMap[colHeader] = point	
			}
			
		}
		tableList = append(tableList, rowMap)
		result, err := json.Marshal(rowMap)
		if err != nil {
			return result, err
		}
	}

	return json.Marshal(tableList)
}

func (t *KeyValueTable) MarshalJSON() ([]byte, error) {
	// Create a list of the rows
	rowMap := make(map[string]interface{})
	var key string 
	var value string
	var js json.RawMessage

	for _, row := range t.rows {
		// Create a map for to represent a row
		key, value = "" , ""
		
		if len(row) == 2 {
			key = row[0]
			value = row[1]
			// if row[1] looks like a JSON string, we need to unmarshal it so it re-marshalls properly.
			if json.Unmarshal([]byte(value), &js) == nil {
				rowMap[key] = js
				// skip the rest of the loop
				continue
			}
		} else if len(row) == 1 {
			key = row[0]
		}
		rowMap[key] = value


	}

	return json.Marshal(rowMap)
}