package codeforces

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	// 正则匹配: | 1 | **4A Watermelon** | 800 |
	tableRowRegex = regexp.MustCompile(`^\s*\|\s*(\d+)\s*\|\s*\*\*(\d+)([A-Za-z]\d*)\s+(.*?)\*\*\s*\|\s*(\d+)\s*\|`)
)

// ParseMarkdownTable 将 Markdown 表格格式的题单清洗为结构化的 ProblemItem 列表
func ParseMarkdownTable(tableContent string) ([]ProblemItem, error) {
	lines := strings.Split(tableContent, "\n")
	var problems []ProblemItem

	for _, line := range lines {
		matches := tableRowRegex.FindStringSubmatch(line)
		if len(matches) == 6 {
			id, _ := strconv.Atoi(matches[1])
			contestID, _ := strconv.Atoi(matches[2])
			index := matches[3]
			name := strings.TrimSpace(matches[4])
			rating, _ := strconv.Atoi(matches[5])

			code := fmt.Sprintf("%d%s", contestID, index)
			url := fmt.Sprintf("https://codeforces.com/problemset/problem/%d/%s", contestID, index)

			problems = append(problems, ProblemItem{
				ID:        id,
				Code:      code,
				ContestID: contestID,
				Index:     index,
				Name:      name,
				Rating:    rating,
				URL:       url,
			})
		}
	}

	return problems, nil
}

// RawDefaultMarkdownTable 用户提供的 60 道 Codeforces 经典题目初始底稿
const RawDefaultMarkdownTable = `|  # | 题目                                       |  难度 |
| -: | ---------------------------------------- | --: |
|  1 | **4A Watermelon**                        | 800 |
|  2 | **71A Way Too Long Words**               | 800 |
|  3 | **231A Team**                            | 800 |
|  4 | **282A Bit++**                           | 800 |
|  5 | **158A Next Round**                      | 800 |
|  6 | **50A Domino piling**                    | 800 |
|  7 | **263A Beautiful Matrix**                | 800 |
|  8 | **112A Petya and Strings**               | 800 |
|  9 | **236A Boy or Girl**                     | 800 |
| 10 | **339A Helpful Maths**                   | 800 |
| 11 | **281A Word Capitalization**             | 800 |
| 12 | **791A Bear and Big Brother**            | 800 |
| 13 | **617A Elephant**                        | 800 |
| 14 | **266A Stones on the Table**             | 800 |
| 15 | **546A Soldier and Bananas**             | 800 |
| 16 | **59A Word**                             | 800 |
| 17 | **977A Wrong Subtraction**               | 800 |
| 18 | **110A Nearly Lucky Number**             | 800 |
| 19 | **734A Anton and Danik**                 | 800 |
| 20 | **41A Translation**                      | 800 |
| 21 | **677A Vanya and Fence**                 | 800 |
| 22 | **271A Beautiful Year**                  | 800 |
| 23 | **116A Tram**                            | 800 |
| 24 | **266B Queue at the School**             | 800 |
| 25 | **469A I Wanna Be the Guy**              | 800 |
| 26 | **486A Calculating Function**            | 800 |
| 27 | **492A Vanya and Cubes**                 | 800 |
| 28 | **520A Pangram**                         | 800 |
| 29 | **581A Vasya the Hipster**               | 800 |
| 30 | **702A Maximum Increase**                | 800 |
| 31 | **710A King Moves**                      | 800 |
| 32 | **749A Bachgold Problem**                | 800 |
| 33 | **978A Remove Duplicates**               | 800 |
| 34 | **988A Diverse Team**                    | 800 |
| 35 | **996A Hit the Lottery**                 | 800 |
| 36 | **1003A Polycarp's Pockets**             | 800 |
| 37 | **1005A Tanya and Stairways**            | 800 |
| 38 | **1030A In Search of an Easy Problem**   | 800 |
| 39 | **1047A Little C Loves 3 I**             | 800 |
| 40 | **1061A Coins**                          | 800 |
| 41 | **1080A Petya and Origami**              | 800 |
| 42 | **1096A Find Divisible**                 | 800 |
| 43 | **1102A Integer Sequence Dividing**      | 800 |
| 44 | **1118A Water Buying**                   | 800 |
| 45 | **1154A Restoring Three Numbers**        | 800 |
| 46 | **1176A Divide it!**                     | 800 |
| 47 | **1183A Nearest Interesting Number**     | 800 |
| 48 | **1207A There Are Two Types Of Burgers** | 800 |
| 49 | **1228A Distinct Digits**                | 800 |
| 50 | **1230A Dawid and Bags of Candies**      | 800 |
| 51 | **1283A Minutes Before the New Year**    | 800 |
| 52 | **1296A Array with Odd Sum**             | 800 |
| 53 | **1323A Even Subset Sum Problem**        | 800 |
| 54 | **1325A EhAb AnD gCd**                   | 800 |
| 55 | **1328A Divisibility Problem**           | 800 |
| 56 | **1352A Sum of Round Numbers**           | 800 |
| 57 | **1370A Maximum GCD**                    | 800 |
| 58 | **1389A LCM Problem**                    | 800 |
| 59 | **1421A XORwice**                        | 800 |
| 60 | **1472B Fair Division**                  | 800 |`

// GetDefaultProblemItems 返回清洗好的 60 道默认题目
func GetDefaultProblemItems() []ProblemItem {
	items, _ := ParseMarkdownTable(RawDefaultMarkdownTable)
	return items
}
