package browscap_go

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"sort"
	"strings"
)

// these constants vere tuned experimentally, dont change them without a good reason
const NGRAM_LEN = 5
const MAX_INDEXLIST = 5000

var (
	// Ini
	sEmpty   = []byte{}     // empty signal
	nComment = []byte{'#'}  // number signal
	sComment = []byte{';'}  // semicolon signal
	sStart   = []byte{'['}  // section start signal
	sEnd     = []byte{']'}  // section end signal
	sEqual   = []byte{'='}  // equal signal
	sQuote1  = []byte{'"'}  // quote " signal
	sQuote2  = []byte{'\''} // quote ' signal

	// To reduce memory usage we will keep only next keys
	keepKeys = [][]byte{
		// Required
		[]byte("Parent"),

		// Used in Browser
		[]byte("Browser"),
		[]byte("Version"),
		[]byte("MajorVer"),
		[]byte("MinorVer"),
		[]byte("Browser_Type"),
		[]byte("Platform"),
		[]byte("Platform_Version"),
		[]byte("Device_Type"),
		//[]byte("Device_Code_Name"),
		//[]byte("Device_Brand_Name"),

		[]byte("Comment"),
		[]byte("Browser_Bits"),
		//[]byte("Browser_Maker"),
		[]byte("Browser_Modus"),
		//[]byte("Platform_Description"),
		[]byte("Platform_Bits"),
		//[]byte("Platform_Maker"),
		[]byte("Alpha"),
		[]byte("Beta"),
		//[]byte("Win16"),
		//[]byte("Win32"),
		//[]byte("Win64"),
		[]byte("Frames"),
		[]byte("IFrames"),
		[]byte("Tables"),
		[]byte("Cookies"),
		[]byte("BackgroundSounds"),
		[]byte("JavaScript"),
		[]byte("VBScript"),
		[]byte("JavaApplets"),
		[]byte("ActiveXControls"),
		//[]byte("isMobileDevice"),
		//[]byte("isTablet"),
		//[]byte("isSyndicationReader"),
		[]byte("Crawler"),
		[]byte("CssVersion"),
		[]byte("AolVersion"),
		[]byte("Device_Name"),
		//[]byte("Device_Maker"),
		[]byte("RenderingEngine_Name"),
		[]byte("RenderingEngine_Version"),
		//[]byte("RenderingEngine_Description"),
		//[]byte("RenderingEngine_Maker"),
	}
)

func loadFromIniFile(path string) (*dictionary, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	dict := newDictionary()

	buf := bufio.NewReader(file)
	sectionName := ""
	sectionPrefix := ""

	for {
		line, _, err := buf.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}

		// Empty line
		if bytes.Equal(sEmpty, line) {
			continue
		}

		// Trim
		line = bytes.TrimSpace(line)

		// Empty line
		if bytes.Equal(sEmpty, line) {
			continue
		}

		// Comment line
		if bytes.HasPrefix(line, nComment) || bytes.HasPrefix(line, sComment) {
			continue
		}

		// Section line
		if bytes.HasPrefix(line, sStart) && bytes.HasSuffix(line, sEnd) {
			sectionName = string(line[1 : len(line)-1])
			sectionPrefix = getPrefix(sectionName)
			continue
		}

		// Create section
		if _, ok := dict.mapped[sectionName]; !ok {
			// Save mapped
			dict.mapped[sectionName] = make(section)

			// Create prefix for section
			if _, exists := dict.expressions[sectionPrefix]; !exists {
				dict.expressions[sectionPrefix] = []*expression{}
			}

			// Build expression
			var ee *expression
			if bytes.IndexAny([]byte(sectionName), "*?") != -1 {
				ee = newRegexpExpression(sectionName)
			} else {
				ee = newCompareExpression(sectionName)
			}
			dict.expressions[sectionPrefix] = append(dict.expressions[sectionPrefix], ee)

			// add the expression to the expression list
			i := len(dict.expressionList)
			dict.expressionList = append(dict.expressionList, ee)
			dict.expressionLengths = append(dict.expressionLengths, float64(len(ee.Name)))
			for _, ngram := range getNGrams(strings.ToLower(sectionName), NGRAM_LEN) {
				if !strings.Contains(ngram, "*") && !strings.Contains(ngram, "?") {
					expressionIndex, found := dict.ngramIndex[ngram]
					if !found {
						expressionIndex = make([]int, 0, 0)
					}
					expressionIndex = append(expressionIndex, i)
					dict.ngramIndex[ngram] = expressionIndex
				}
			}

		}

		// Key => Value
		kv := bytes.SplitN(line, sEqual, 2)

		// Parse Key
		key := bytes.TrimSpace(kv[0])
		if !inList(key, keepKeys) {
			continue
		}

		// Parse Value
		val := bytes.TrimSpace(kv[1])
		if bytes.HasPrefix(val, sQuote1) {
			val = bytes.Trim(val, `"`)
		}
		if bytes.HasPrefix(val, sQuote2) {
			val = bytes.Trim(val, `'`)
		}

		dict.mapped[sectionName][string(key)] = string(val)
	}

	// delete mgrams which are very common
	toDelete := make(map[string]struct{})
	for ngram, indexList := range dict.ngramIndex {
		if len(indexList) > MAX_INDEXLIST {
			toDelete[ngram] = struct{}{}
		}
	}
	for ngram, _ := range toDelete {
		delete(dict.ngramIndex, ngram)
	}

	// Order expressions by length
	for prefix := range dict.expressions {
		expressions := dict.expressions[prefix]
		sort.Sort(expressionByNameLen(expressions))
		dict.expressions[prefix] = expressions
	}

	return dict, nil
}
