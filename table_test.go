package table_test

import (
	"bytes"
	"testing"
	"github.com/stretchr/testify/assert"

	. "github.com/allmightyspiff/terminal-table"
)

// Happy path testing
func TestPrintTableSimple(t *testing.T) {
	buf := bytes.Buffer{}
	testTable := NewTable([]string{"test1", "test2"})
	testTable.SetOutput(&buf)
	testTable.Add("row1", "row2")
	testTable.Print()
	assert.Contains(t, buf.String(), "test2")
	assert.Contains(t, buf.String(), "row1")
	assert.Equal(t, "test1   test2\nrow1    row2\n", buf.String())
}


func TestPrintTableJson(t *testing.T) {
	buf := bytes.Buffer{}
	testTable := NewTable([]string{"test1", "test2"})
	testTable.SetOutput(&buf)
	testTable.Add("row1-col1", "row1-col2")
	testTable.Add("row2-col1", "row2-col2")
	testTable.PrintJson()
	assert.Contains(t, buf.String(), "\"test1\": \"row1-col1\"")
	assert.Contains(t, buf.String(), "\"test2\": \"row2-col2\"")
}

// Blank headers
func TestEmptyHeaderTable(t *testing.T) {
	buf := bytes.Buffer{}
	testTable := NewTable([]string{"", ""})
	testTable.SetOutput(&buf)
	testTable.Add("row1", "row2")
	testTable.Print()
	assert.Contains(t, buf.String(), "row1")
	assert.Equal(t, "       \nrow1   row2\n", buf.String())
}


func TestEmptyHeaderTableJson(t *testing.T) {
	buf := bytes.Buffer{}
	testTable := NewTable([]string{"", ""})
	testTable.SetOutput(&buf)
	testTable.Add("row1", "row2")
	testTable.PrintJson()

	assert.Contains(t, buf.String(), "\"column_2\": \"row2\"")
	assert.Contains(t, buf.String(), "\"column_1\": \"row1\"")
}

// Empty Headers / More rows than headers
func TestZeroHeadersTable(t *testing.T) {
	buf := bytes.Buffer{}
	testTable := NewTable([]string{})
	testTable.SetOutput(&buf)
	testTable.Add("row1", "row2")
	testTable.Print()
	assert.Contains(t, buf.String(), "row1")
	assert.Equal(t, "\nrow1   row2\n", buf.String())
}

func TestZeroHeadersTableJson(t *testing.T) {
	buf := bytes.Buffer{}
	testTable := NewTable([]string{})
	testTable.SetOutput(&buf)
	testTable.Add("row1", "row2")
	testTable.PrintJson()
	assert.Contains(t, buf.String(), "row1")
	assert.Contains(t, buf.String(), "\"column_2\": \"row2\"")
	assert.Contains(t, buf.String(), "\"column_1\": \"row1\"")
}

// Empty rows / More headers than rows

func TestNotEnoughRowEntires(t *testing.T) {
	buf := bytes.Buffer{}
	testTable := NewTable([]string{"col1", "col2"})
	testTable.SetOutput(&buf)
	testTable.Add("row1")
	testTable.Add("", "row2")
	testTable.Print()
	assert.Contains(t, buf.String(), "row1")
	assert.Equal(t, "col1   col2\nrow1   \n       row2\n", buf.String())
}

func TestNotEnoughRowEntiresJson(t *testing.T) {
	buf := bytes.Buffer{}
	testTable := NewTable([]string{})
	testTable.SetOutput(&buf)
	testTable.Add("row1")
	testTable.Add("", "row2")
	testTable.PrintJson()
	assert.Contains(t, buf.String(), "row1")
	assert.Contains(t, buf.String(), "\"column_2\": \"row2\"")
	assert.Contains(t, buf.String(), "\"column_1\": \"row1\"")
	assert.Contains(t, buf.String(), "\"column_1\": \"\"")
}



// Old way of creating sub-tables
// Key        Value
// FirstKey   FirstValue
// SubTable   AAAA      BBBB
//            Sub-AA1   Sub-BB1
//            Sub-AA2   Sub-BB2

// MIght want to drop support for this one?
func TestNestedAndTextOutput1(t *testing.T) {
	buf := bytes.Buffer{}
	subBuf := bytes.Buffer{}
	mainTable := NewTable([]string{"Key", "Value"})
	mainTable.SetOutput(&buf)
	mainTable.Add("FirstKey", "FirstValue")
	subTable := NewTable([]string{"AAAA", "BBBB"})
	subTable.SetOutput(&subBuf)
	subTable.Add("Sub-AA1", "Sub-BB1")
	subTable.Add("Sub-AA2", "Sub-BB2")
	subTable.Print()
	mainTable.Add("SubTable", subBuf.String())
	mainTable.Print()
	theOutput := buf.String()
	assert.Contains(t, theOutput, "FirstKey   FirstValue")
	assert.Contains(t, theOutput, "SubTable   AAAA      BBBB")
	assert.Contains(t, theOutput, "           Sub-AA2   Sub-BB2")

}



// New way of creating sub-tables
// Key        Value
// FirstKey   FirstValue
// SubTable   AAAA      BBBB
//            Sub-AA1   Sub-BB1
//            Sub-AA2   Sub-BB2

func TestNestedAndTextOutput2(t *testing.T) {
	outBuf := bytes.Buffer{}
	mainTable := NewTable([]string{"Key", "Value"})
	mainTable.SetOutput(&outBuf)
	mainTable.Add("FirstKey", "FirstValue")

	subTable := NewTable([]string{"AAAA", "BBBB"})
	subTable.Add("Sub-AA1", "Sub-BB1")
	subTable.Add("Sub-AA2", "Sub-BB2")
	mainTable.AddNestedTable("SubTable", subTable)
	mainTable.Print()
	theOutput := outBuf.String()
	assert.Contains(t, theOutput, "FirstKey   FirstValue")
	assert.Contains(t, theOutput, "SubTable   AAAA      BBBB")
	assert.Contains(t, theOutput, "           Sub-AA2   Sub-BB2")

}

// New way of creating sub-tables
// "[{"HOne":"Row1-1","HThree":"AAAA      BBBB","HTwo":"99"},{"HOne":"","HThree":"SUB-AAA   SUB-BBB","HTwo":""},{"HOne":"","HThree":"","HTwo":""}]

func TestNestedAndJsonOutput2(t *testing.T) {
	outBuf := bytes.Buffer{}

	mainTable := NewTable([]string{"HOne", "HTwo", "HThree"})
	mainTable.SetOutput(&outBuf)
	subTable := NewTable([]string{"AAAA", "BBBB"})
	subTable.SetFormat("Json")
	mainTable.SetFormat("json")
	subTable.Add("SUB-AAA", "SUB-BBB")
	mainTable.AddNestedTable("Row1-1", 99, subTable)
	
	mainTable.PrintJson()
	theOutput := outBuf.String()
	assert.Contains(t, theOutput, `"HTwo": 99`)
	assert.Contains(t, theOutput, `"HThree": [`)
	assert.Contains(t, theOutput, `"AAAA": "SUB-AAA",`)


}


// Nested tables with JSON output.
///{
//         "FirstKey": "FirstValue",
//         "SubTable": [
//         {
//                 "AAAA": "Sub-AA1",
//                 "BBBB": "Sub-BB1"
//         },
//         {
//                 "AAAA": "Sub-AA2",
//                 "BBBB": "Sub-BB2"
//         }
// ]
// }
func TestNestedAndJsonOutput1(t *testing.T) {
	outBuf := bytes.Buffer{}
	mainTable := NewTable([]string{"Key", "Value"})
	mainTable.SetFormat("JSON")
	mainTable.SetOutput(&outBuf)
	mainTable.Add("FirstKey", "FirstValue")

	subTable := NewTable([]string{"AAAA", "BBBB"})
	subTable.SetFormat("Json")
	subTable.Add("Sub-AA1", "Sub-BB1")
	subTable.Add("Sub-AA2", "Sub-BB2")
	mainTable.AddNestedTable("SubTable", subTable)
	mainTable.PrintJson()
	theOutput := outBuf.String()
	assert.Contains(t, theOutput, `"FirstKey": "FirstValue",`)
	assert.Contains(t, theOutput, `"SubTable": [`)
	assert.Contains(t, theOutput, `"AAAA": "Sub-AA1",`)

}

func TestEllipsisTable(t *testing.T) {
	headers := []string{"A", "B", "C"}
	outBuf := bytes.Buffer{}
	mainTable := NewTable(headers)
	mainTable.SetOutput(&outBuf)
	row := make([]string, len(headers))
	row[0] = "1aaa"
	row[1] = "1bbb"
	row[2] = "1ccc"
	mainTable.Add(row...)
	mainTable.Print()
	theOutput := outBuf.String()
	assert.Contains(t, theOutput, "1aaa   1bbb   1ccc")
}


func TestSLSubnetListProblem(t *testing.T) {

	headers := []string{"QQQ", "AAA", "ZZZ"}

	outBuf := bytes.Buffer{}
	mainTable := NewTable(headers)
	mainTable.SetOutput(&outBuf)
	mainTable.SetFormat("json")
	mainTable.Add("10","123","0")
	mainTable.Add("10","123","10")
	mainTable.PrintJson()
	theOutput := outBuf.String()
	assert.Contains(t, theOutput, `"ZZZ": 0`)
	assert.Contains(t, theOutput, `"ZZZ": 10`)

}
