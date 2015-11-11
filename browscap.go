package browscap_go

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"
)

const (
	DownloadUrl     = "http://browscap.org/stream?q=PHP_BrowsCapINI"
	CheckVersionUrl = "http://browscap.org/version-number"
)

var (
	dict        *dictionary
	initialized bool
	version     string
	debug       bool
)

func Debug(val bool) {
	debug = val
}

func InitBrowsCap(path string, force bool) error {
	if initialized && !force {
		return nil
	}
	var err error

	// Load ini file
	if dict, err = loadFromIniFile(path); err != nil {
		return fmt.Errorf("browscap: An error occurred while reading file, %v ", err)
	}

	if verDictionary, exists := dict.mapped["GJK_Browscap_Version"]; exists {
		version = verDictionary["Version"]
	}

	initialized = true
	return nil
}

func InitializedVersion() string {
	return version
}

func LastVersion() (string, error) {
	response, err := http.Get(CheckVersionUrl)
	if err != nil {
		return "", fmt.Errorf("browscap: error sending request, %v", err)
	}
	defer response.Body.Close()

	// Get body of response
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("browscap: error reading the response data of request, %v", err)
	}

	// Check 200 status
	if response.StatusCode != 200 {
		return "", fmt.Errorf("browscap: error unexpected status code %d", response.StatusCode)
	}

	return string(body), nil
}

func DownloadFile(saveAs string) error {
	response, err := http.Get(DownloadUrl)
	if err != nil {
		return fmt.Errorf("browscap: error sending request, %v", err)
	}
	defer response.Body.Close()

	// Get body of response
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("browscap: error reading the response data of request, %v", err)
	}

	// Check 200 status
	if response.StatusCode != 200 {
		return fmt.Errorf("browscap: error unexpected status code %d", response.StatusCode)
	}

	if err = ioutil.WriteFile(saveAs, body, os.ModePerm); err != nil {
		return fmt.Errorf("browscap: error saving file, %v", err)
	}

	return nil
}

func GetBrowserData(userAgent string) (map[string]string, bool) {
	bdata, ok := searchIndexedBrowserData(userAgent)
	if ok {
		return bdata, ok
	}
	return getLoopBrowserData(userAgent)
}

func getLoopBrowserData(userAgent string) (bdata map[string]string, ok bool) {
	if !initialized {
		return
	}

	agent := bytes.ToLower([]byte(userAgent))
	prefix := getPrefix(userAgent)

	// Main search
	if bdata, ok = getBrowserData(prefix, agent); ok {
		return
	}

	// Fallback
	if prefix != "*" {
		bdata, ok = getBrowserData("*", agent)
	}

	return
}

func searchIndexedBrowserData(userAgent string) (map[string]string, bool) {
	if !initialized {
		return nil, false
	}

	agent := strings.ToLower(userAgent)
	agentBytes := []byte(agent)

	nonemptyLists := make(hitPairListList, 0)

	for _, ngram := range getNGrams(agent, NGRAM_LEN) {
		eeIxList, ok := dict.ngramIndex[ngram]
		if ok && len(eeIxList) > 0 {
			nonemptyLists = append(nonemptyLists, eeIxList)
		}
	}

	sort.Sort(nonemptyLists) // shorter first

	listIdx := make([]int, len(nonemptyLists))

	for startI := 0; startI < len(nonemptyLists); {
		for i := startI; i < len(listIdx); i++ {
			idx := listIdx[i]
			if idx >= len(nonemptyLists[i]) {
				startI++
				continue
			}
			ee := dict.expressionList[nonemptyLists[i][idx].Key]
			if ee.Match(agentBytes) {
				data := dict.findData(ee.Name)
				return data, true
			}
			listIdx[i]++
		}
	}

	return nil, false
}

type hitPair struct {
	Key int
	Val float64
}

type hitPairList []hitPair

func (p hitPairList) Len() int { return len(p) }
func (p hitPairList) Less(i, j int) bool {
	return p[i].Val < p[j].Val
}
func (p hitPairList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

type hitPairListList []hitPairList

func (p hitPairListList) Len() int { return len(p) }
func (p hitPairListList) Less(i, j int) bool {
	return len(p[i]) < len(p[j])
}
func (p hitPairListList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func getBrowserData(prefix string, agent []byte) (map[string]string, bool) {
	if expressions, exists := dict.expressions[prefix]; exists {
		for _, exp := range expressions {
			if exp.Match(agent) {
				data := dict.findData(exp.Name)
				return data, true
			}
		}
	}
	return map[string]string{}, false
}
