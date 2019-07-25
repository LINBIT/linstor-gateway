// Template processor
//
// Replaces variables in text templates.
// Currently used to create XML documents or parts thereof,
// but can be used for any line-based text input.
package templateproc

import "os"
import "io"
import "bufio"
import "fmt"

const (
	RPL_COPY = iota
	RPL_ESCAPED
	RPL_VAR_SIGN
	RPL_VAR_NAME
)

// Loads a template from a file into a string array, one line per array element
func LoadTemplate(fileName string) ([]string, error) {
	var tmplLines []string

	tmplFile, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer tmplFile.Close()

	tmplReader := bufio.NewReader(tmplFile)
	err = nil
	for {
		line, err := tmplReader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}
		tmplLines = append(tmplLines, line)
	}

	return tmplLines, nil
}

// Replaces variable keys with variable values in a template.
//
// Returns: the text created by replacing the variables in the template
func ReplaceVariables(tmplLines []string, varMap map[string]string) []string {
	var resultLines []string
	lineCount := len(tmplLines)
	for lineIdx := 0; lineIdx < lineCount; lineIdx++ {
		var outLine []byte
		offset := 0
		state := RPL_COPY
		lineLength := len(tmplLines[lineIdx])
		for byteIdx := 0; byteIdx < lineLength; byteIdx++ {
			switch state {
			case RPL_ESCAPED:
				state = RPL_COPY
				outLine = append(outLine, tmplLines[lineIdx][offset:byteIdx]...)
				offset = byteIdx + 1
			case RPL_VAR_SIGN:
				if tmplLines[lineIdx][byteIdx] == '{' {
					state = RPL_VAR_NAME
					offset = byteIdx + 1
				} else {
					state = RPL_COPY
					outLine = append(outLine, '$')
					outLine = append(outLine, '{')
					offset = byteIdx + 1
				}
			case RPL_VAR_NAME:
				if tmplLines[lineIdx][byteIdx] == '}' {
					state = RPL_COPY
					varName := string(tmplLines[lineIdx][offset:byteIdx])
					offset = byteIdx + 1

					varValue, foundFlag := varMap[varName]
					if foundFlag {
						outLine = append(outLine, varValue...)
					} else {
						fmt.Printf("\x1b[1;33mWarning: Template contains undefined variable '%s'\x1b[0m\n", varName)
					}
				}
			case RPL_COPY:
				fallthrough
			default:
				if tmplLines[lineIdx][byteIdx] == '\\' {
					outLine = append(outLine, tmplLines[lineIdx][offset:byteIdx]...)
					offset = byteIdx + 1
					state = RPL_ESCAPED
				} else if tmplLines[lineIdx][byteIdx] == '$' {
					outLine = append(outLine, tmplLines[lineIdx][offset:byteIdx]...)
					offset = byteIdx + 1
					state = RPL_VAR_SIGN
				}
			}
		}
		outLine = append(outLine, tmplLines[lineIdx][offset:lineLength]...)
		resultLines = append(resultLines, string(outLine))
	}
	return resultLines
}
