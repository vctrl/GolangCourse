package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// вам надо написать более быструю оптимальную этой функции
func SolutionSearch(out io.Writer) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	s := bufio.NewScanner(file)

	buf := make([]byte, 4096*16)
	s.Buffer(buf, 566019)

	if err != nil {
		panic(err)
	}

	uniqueBrowsers := make(map[string]struct{})
	var foundUsersSb strings.Builder
	foundUsersSb.Grow(4188)

	i := -1
	for s.Scan() {
		i++
		var user User
		// fmt.Printf("%v %v\n", err, line)
		bytes := s.Bytes()
		err := user.UnmarshalJSON(bytes)
		if err != nil {
			panic(err)
		}

		isAndroid := false
		isMSIE := false

		for _, browser := range user.Browsers {
			currAndroid := strings.Contains(browser, "Android")
			currMSIE := strings.Contains(browser, "MSIE")
			if currAndroid {
				isAndroid = true
			}
			if currMSIE {
				isMSIE = true
			}

			if currAndroid || currMSIE {
				uniqueBrowsers[browser] = struct{}{}
			}
		}

		if !(isAndroid && isMSIE) {
			continue
		}

		// log.Println("Android and MSIE user:", user["name"], user["email"])
		email := strings.Replace(user.Email, "@", " [at] ", 1)
		foundUsersSb.Write([]byte(fmt.Sprintf("[%d] %s <%s>\n", i, user.Name, email)))
	}

	fmt.Fprintln(out, "found users:\n"+foundUsersSb.String())
	fmt.Fprintln(out, "Total unique browsers", len(uniqueBrowsers))
}
