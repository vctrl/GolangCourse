package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	goodToken = "token"
	badToken  = "unknown"
)

type User1 struct {
	ID        int    `xml:"id"`
	FirstName string `xml:"first_name"`
	LastName  string `xml:"last_name"`
	Age       int    `xml:"age"`
	About     string `xml:"about"`
	Gender    string `xml:"gender"`
}

type UsersXML []User1

func (us *UsersXML) MapToUsers() []User {
	arr := []User1(*us)
	res := make([]User, len(arr))

	for i, u := range arr {
		res[i] = User{Id: u.ID, Name: u.FirstName + " " + u.LastName, Age: u.Age, About: u.About, Gender: u.Gender}
	}

	return res
}

type Users struct {
	XMLName xml.Name `xml:"root"`
	Users   []User1  `xml:"row"`
}

type ErrorCase struct {
	HandlerFunc func(w http.ResponseWriter, r *http.Request)
	Result      string
}

type Case struct {
	HandlerFunc func(w http.ResponseWriter, r *http.Request)
	Limit       int
	Offset      int
	OrderBy     int
	Query       string
	OrderField  string
	Result      string
}

var errCases = []ErrorCase{
	{
		HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(time.Second * 1000)
		},
		Result: "timeout for limit=1&offset=0&order_by=0&order_field=&query=",
	},
	{
		HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		},
		Result: "SearchServer fatal error",
	},
	{
		HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "{")
		},
		Result: fmt.Sprintf("cant unpack error json: %s", "unexpected end of JSON input"),
	},
	{
		HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
			respErrJSON, _ := json.Marshal(&SearchErrorResponse{Error: "Unknown error"})
			w.WriteHeader(http.StatusBadRequest)
			w.Write(respErrJSON)
		},
		Result: fmt.Sprintf("unknown bad request error: %s", "Unknown error"),
	},
	{
		HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "%s", "{")
		},
		Result: fmt.Sprintf("cant unpack result json: unexpected end of JSON input"),
	},
}

func TestErrors(t *testing.T) {
	for i, test := range errCases {
		_, err := SendTestRequest(test.HandlerFunc, 0, 0, OrderByAsIs, "", "")
		if err == nil {
			t.Fatalf("Error case %d failed: Expected error, have nil", i)
		}

		if err.Error() != test.Result {
			t.Fatalf("case %d failed: expected %s, got %v", i, test.Result, err)
		}
	}
}

var cases = []Case{
	{
		HandlerFunc: SearchServer,
		Limit:       -1,
		Offset:      0,
		OrderBy:     OrderByAsIs,
		Query:       "",
		OrderField:  "",
		Result:      "limit must be > 0",
	},
	{
		HandlerFunc: SearchServer,
		Limit:       0,
		Offset:      -1,
		OrderBy:     OrderByAsIs,
		Query:       "",
		OrderField:  "",
		Result:      "offset must be > 0",
	},
	{
		HandlerFunc: SearchServer,
		Limit:       0,
		Offset:      -1,
		OrderBy:     OrderByAsIs,
		Query:       "",
		OrderField:  "",
		Result:      "offset must be > 0",
	},
	{
		HandlerFunc: SearchServer,
		Limit:       0,
		Offset:      0,
		OrderBy:     OrderByAsIs,
		Query:       "",
		OrderField:  "unknown",
		Result:      "OrderFeld unknown invalid",
	},
}

func TestParams(t *testing.T) {
	for i, test := range cases {
		_, err := SendTestRequest(SearchServer, test.Limit, test.Offset, test.OrderBy, test.Query, test.OrderField)

		if err == nil {
			t.Fatalf("Error case %d failed: expacted error on negative limit value, have nil", i)
		}

		if err.Error() != test.Result {
			t.Errorf("case %d failed: unexpected error message, expected: %s, but was: %s", i, test.Result, err.Error())
		}
	}

}

func SearchServerUnknownBadRequestError(w http.ResponseWriter, r *http.Request) {
	respErrJSON, _ := json.Marshal(&SearchErrorResponse{Error: "Unknown error"})
	w.WriteHeader(http.StatusBadRequest)
	w.Write(respErrJSON)
}

func SearchServerSuccessInvalidJSON(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s", "{")
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("AccessToken") != goodToken {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	xmlFile, err := os.Open("dataset.xml")
	if err != nil {
		fmt.Printf("error while opening file: %v\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	byteValue, err := ioutil.ReadAll(xmlFile)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var users Users

	err = xml.Unmarshal(byteValue, &users)
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	queryParam := r.FormValue("query")
	queried := query(users.Users, queryParam)

	orderFieldParam := r.FormValue("order_field")

	if orderFieldParam != "Id" && orderFieldParam != "Age" &&
		orderFieldParam != "Name" && orderFieldParam != "" {
		respErrJSON, _ := json.Marshal(&SearchErrorResponse{Error: "ErrorBadOrderField"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(respErrJSON)

		return
	}

	orderByParam, err := strconv.Atoi(r.FormValue("order_by"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if orderByParam != OrderByAsIs {
		sortUsers(queried, orderFieldParam, orderByParam)
	}

	limit, err := strconv.Atoi(r.FormValue("limit"))

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	offset, err := strconv.Atoi(r.FormValue("offset"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var result UsersXML
	if offset >= len(queried) {
		result = UsersXML{}
	} else if offset+limit > len(queried) {
		result = queried[offset:]
	} else {
		result = UsersXML(queried[offset : offset+limit])
	}

	// for _, u := range result {
	// 	fmt.Printf("&User{Id: %d, Name: \"%s\", Age: %d, About: \"%s\", Gender:\"%s\"}\n", u.ID, u.FirstName+" "+u.LastName, u.Age, u.About, u.Gender)
	// }
	resultJSON, err := json.Marshal(result.MapToUsers())

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(resultJSON)
}

func query(users []User1, param string) []User1 {
	res := make([]User1, 0)
	for _, u := range users {
		if strings.Contains(u.FirstName, param) ||
			strings.Contains(u.LastName, param) ||
			strings.Contains(u.About, param) {
			res = append(res, u)
		}
	}

	return res
}

func sortUsers(users []User1, orderField string, orderBy int) {
	intComparator := getIntComparator(orderBy)
	strComparator := getStringComparator(orderBy)

	sort.Slice(users, func(i, j int) bool {
		switch orderField {
		case "Id":
			return intComparator(users[i].ID, users[j].ID)
		case "Age":
			return intComparator(users[i].Age, users[j].Age)
		default:
			if users[i].FirstName == users[j].FirstName {
				return strComparator(users[i].LastName, users[j].LastName)
			}

			return strComparator(users[i].FirstName, users[j].FirstName)
		}
	})
}

func getIntComparator(orderBy int) func(i, j int) bool {
	var comparator func(i, j int) bool
	switch orderBy {
	case OrderByAsc:
		comparator = func(i, j int) bool {
			return i < j
		}
	case OrderByDesc:
		comparator = func(i, j int) bool {
			return i > j
		}
	}

	return comparator
}

// где-то здесь тоска по дженерикам
func getStringComparator(orderBy int) func(i, j string) bool {
	var comparator func(i, j string) bool
	switch orderBy {
	case OrderByAsc:
		comparator = func(i, j string) bool {
			return i < j
		}
	case OrderByDesc:
		comparator = func(i, j string) bool {
			return i > j
		}
	}

	return comparator
}

func TestLimitHappyCase(t *testing.T) {
	result, err := SendTestRequest(SearchServer, 6, 0, OrderByAsIs, "", "")
	if err != nil {
		fmt.Println(err)
	}

	expectedResult := []User{
		{Id: 0, Name: "Boyd Wolf", Age: 22, About: "Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.\n", Gender: "male"},
		{Id: 1, Name: "Hilda Mayer", Age: 21, About: "Sit commodo consectetur minim amet ex. Elit aute mollit fugiat labore sint ipsum dolor cupidatat qui reprehenderit. Eu nisi in exercitation culpa sint aliqua nulla nulla proident eu. Nisi reprehenderit anim cupidatat dolor incididunt laboris mollit magna commodo ex. Cupidatat sit id aliqua amet nisi et voluptate voluptate commodo ex eiusmod et nulla velit.\n", Gender: "female"},
		{Id: 2, Name: "Brooks Aguilar", Age: 25, About: "Velit ullamco est aliqua voluptate nisi do. Voluptate magna anim qui cillum aliqua sint veniam reprehenderit consectetur enim. Laborum dolore ut eiusmod ipsum ad anim est do tempor culpa ad do tempor. Nulla id aliqua dolore dolore adipisicing.\n", Gender: "male"},
		{Id: 3, Name: "Everett Dillard", Age: 27, About: "Sint eu id sint irure officia amet cillum. Amet consectetur enim mollit culpa laborum ipsum adipisicing est laboris. Adipisicing fugiat esse dolore aliquip quis laborum aliquip dolore. Pariatur do elit eu nostrud occaecat.\n", Gender: "male"},
		{Id: 4, Name: "Owen Lynn", Age: 30, About: "Elit anim elit eu et deserunt veniam laborum commodo irure nisi ut labore reprehenderit fugiat. Ipsum adipisicing labore ullamco occaecat ut. Ea deserunt ad dolor eiusmod aute non enim adipisicing sit ullamco est ullamco. Elit in proident pariatur elit ullamco quis. Exercitation amet nisi fugiat voluptate esse sit et consequat sit pariatur labore et.\n", Gender: "male"},
		{Id: 5, Name: "Beulah Stark", Age: 30, About: "Enim cillum eu cillum velit labore. In sint esse nulla occaecat voluptate pariatur aliqua aliqua non officia nulla aliqua. Fugiat nostrud irure officia minim cupidatat laborum ad incididunt dolore. Fugiat nostrud eiusmod ex ea nulla commodo. Reprehenderit sint qui anim non ad id adipisicing qui officia Lorem.\n", Gender: "female"},
	}

	if len(result.Users) != len(expectedResult) {
		t.Errorf("results not match\nGot: %v\nExpected: %v", result, expectedResult)
	}

	for i, u := range result.Users {
		if u != expectedResult[i] {
			t.Errorf("results not match on index %d\nGot: %v\nExpected: %v", i, u, expectedResult[i])
		}
	}

	if result.NextPage != true {
		t.Errorf("Expected NextPage value: true, but was: false")
	}
}

// Тесты SearchServer
func TestResultsShouldBeInAscendingOrder(t *testing.T) {
	orderFieldsInt := map[string]func(*User) int{"Id": func(u *User) int { return u.Id }, "Age": func(u *User) int { return u.Age }}
	orderFieldsString := map[string]func(*User) string{"Name": func(u *User) string { return u.Name }, "": func(u *User) string { return u.Name }}

	for f, getValue := range orderFieldsInt {
		result, _ := SendTestRequest(SearchServer, 25, 0, OrderByAsc, "", f)
		checkOrderInt(t, result.Users, getValue, getIntComparator(OrderByAsc))
	}

	for f, getValue := range orderFieldsString {
		result, _ := SendTestRequest(SearchServer, 25, 0, OrderByAsc, "", f)
		checkOrderString(t, result.Users, getValue, getStringComparator(OrderByAsc))
	}
}

func TestResultsShouldBeInDescendingOrder(t *testing.T) {
	orderFieldsInt := map[string]func(*User) int{"Id": func(u *User) int { return u.Id }, "Age": func(u *User) int { return u.Age }}
	orderFieldsString := map[string]func(*User) string{"Name": func(u *User) string { return u.Name }, "": func(u *User) string { return u.Name }}

	for f, getValue := range orderFieldsInt {
		result, _ := SendTestRequest(SearchServer, 25, 0, OrderByDesc, "", f)
		checkOrderInt(t, result.Users, getValue, getIntComparator(OrderByDesc))
	}

	for f, getValue := range orderFieldsString {
		result, _ := SendTestRequest(SearchServer, 25, 0, OrderByDesc, "", f)
		checkOrderString(t, result.Users, getValue, getStringComparator(OrderByDesc))
	}
}

func checkOrderInt(t *testing.T, users []User, getValue func(*User) int, comparator func(int, int) bool) {
	for i := 0; i < len(users)-1; i++ {
		v1 := getValue(&users[i])
		v2 := getValue(&users[i+1])
		if v1 != v2 && !comparator(v1, v2) {
			t.Fatalf("not sorted, indexes: %d %d, values: %v %v", i, i+1, getValue(&users[i]), getValue(&users[i+1]))
		}
	}
}

func checkOrderString(t *testing.T, users []User, getValue func(*User) string, comparator func(string, string) bool) {
	for i := 0; i < len(users)-1; i++ {
		v1 := getValue(&users[i])
		v2 := getValue(&users[i+1])
		if v1 != v2 && !comparator(v1, v2) {
			t.Fatalf("not sorted, indexes: %d %d, values: %v, %v", i, i+1, getValue(&users[i]), getValue(&users[i+1]))
		}
	}
}

func TestResultsShouldBeFilteredByNameOrAbout(t *testing.T) {
	// по Name
	query := "Stark"
	result, _ := SendTestRequest(SearchServer, 25, 0, OrderByDesc, query, "")
	expectedResult := []User{
		{Id: 5, Name: "Beulah Stark", Age: 30, About: "Enim cillum eu cillum velit labore. In sint esse nulla occaecat voluptate pariatur aliqua aliqua non officia nulla aliqua. Fugiat nostrud irure officia minim cupidatat laborum ad incididunt dolore. Fugiat nostrud eiusmod ex ea nulla commodo. Reprehenderit sint qui anim non ad id adipisicing qui officia Lorem.\n", Gender: "female"},
	}

	for i, u := range result.Users {
		if u != expectedResult[i] {
			t.Errorf("results not match on index %d\nGot: %v\nExpected: %v", i, u, expectedResult[i])
		}
	}

	query = "reprehenderit"
	result, _ = SendTestRequest(SearchServer, 25, 0, OrderByDesc, query, "")

	// по About
	for _, u := range result.Users {
		if !strings.Contains(u.About, query) {
			t.Errorf("wrong result for query param %s: %v", query, u)
		}
	}
}

func TestOffset(t *testing.T) {
	result, err := SendTestRequest(SearchServer, 4, 2, OrderByAsIs, "", "")
	if err != nil {
		fmt.Println(err)
	}

	expectedResult := []User{
		{Id: 2, Name: "Brooks Aguilar", Age: 25, About: "Velit ullamco est aliqua voluptate nisi do. Voluptate magna anim qui cillum aliqua sint veniam reprehenderit consectetur enim. Laborum dolore ut eiusmod ipsum ad anim est do tempor culpa ad do tempor. Nulla id aliqua dolore dolore adipisicing.\n", Gender: "male"},
		{Id: 3, Name: "Everett Dillard", Age: 27, About: "Sint eu id sint irure officia amet cillum. Amet consectetur enim mollit culpa laborum ipsum adipisicing est laboris. Adipisicing fugiat esse dolore aliquip quis laborum aliquip dolore. Pariatur do elit eu nostrud occaecat.\n", Gender: "male"},
		{Id: 4, Name: "Owen Lynn", Age: 30, About: "Elit anim elit eu et deserunt veniam laborum commodo irure nisi ut labore reprehenderit fugiat. Ipsum adipisicing labore ullamco occaecat ut. Ea deserunt ad dolor eiusmod aute non enim adipisicing sit ullamco est ullamco. Elit in proident pariatur elit ullamco quis. Exercitation amet nisi fugiat voluptate esse sit et consequat sit pariatur labore et.\n", Gender: "male"},
		{Id: 5, Name: "Beulah Stark", Age: 30, About: "Enim cillum eu cillum velit labore. In sint esse nulla occaecat voluptate pariatur aliqua aliqua non officia nulla aliqua. Fugiat nostrud irure officia minim cupidatat laborum ad incididunt dolore. Fugiat nostrud eiusmod ex ea nulla commodo. Reprehenderit sint qui anim non ad id adipisicing qui officia Lorem.\n", Gender: "female"},
	}

	if len(result.Users) != len(expectedResult) {
		t.Errorf("results not match\nGot: %v\nExpected: %v", result, expectedResult)
	}

	for i, u := range result.Users {
		if u != expectedResult[i] {
			t.Errorf("results not match on index %d\nGot: %v\nExpected: %v", i, u, expectedResult[i])
		}
	}

	if result.NextPage != true {
		t.Errorf("Expected NextPage value: true, but was: false")
	}
}

func TestBadAccessTokenError(t *testing.T) {
	_, err := SendTestRequestBadToken(SearchServer, 0, 0, OrderByAsIs, "", "")
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	if err.Error() != "Bad AccessToken" {
		t.Fatalf("Expected bad access token error, got %v", err)
	}
}

func TestMaxLimitEquals25(t *testing.T) {
	result, err := SendTestRequest(SearchServer, 26, 0, OrderByAsIs, "", "")
	if err != nil {
		t.Fatalf("TestMaxLimitEquals25 fail: unexpected error: %v", err)
	}

	if len(result.Users) > 25 {
		t.Errorf("Unexpected result lenght: %v", len(result.Users))
	}

}

func TestUnknownResponseError(t *testing.T) {
	sc := &SearchClient{URL: ""}

	r := &SearchRequest{
		Limit:      0,
		Offset:     0,
		Query:      "",
		OrderField: "",
		OrderBy:    OrderByAsIs,
	}

	_, err := sc.FindUsers(*r)
	if err == nil {
		t.Fatalf("Expected unsupported protocol scheme error, have nil")
	}
}

func TestResultLenNotEqualLimit(t *testing.T) {
	result, err := SendTestRequest(SearchServer, 6, 0, OrderByAsIs, "Boyd", "")
	if err != nil {
		fmt.Println(err)
	}

	expectedResult := []User{
		{Id: 0, Name: "Boyd Wolf", Age: 22, About: "Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.\n", Gender: "male"},
	}

	if len(result.Users) != len(expectedResult) {
		t.Errorf("results not match\nGot: %v\nExpected: %v", result, expectedResult)
	}

	for i, u := range result.Users {
		if u != expectedResult[i] {
			t.Errorf("results not match on index %d\nGot: %v\nExpected: %v", i, u, expectedResult[i])
		}
	}

	if result.NextPage != false {
		t.Errorf("Expected NextPage value: false, but was: true")
	}
}

func SendTestRequest(handlerFunc http.HandlerFunc, limit, offset, orderBy int, query, orderField string) (*SearchResponse, error) {
	return SendTestRequestInternal(handlerFunc, limit, offset, orderBy, query, orderField, goodToken)
}

func SendTestRequestBadToken(handlerFunc http.HandlerFunc, limit, offset, orderBy int, query, orderField string) (*SearchResponse, error) {
	return SendTestRequestInternal(handlerFunc, limit, offset, orderBy, query, orderField, badToken)
}

func SendTestRequestInternal(handlerFunc http.HandlerFunc, limit, offset, orderBy int, query, orderField, token string) (*SearchResponse, error) {
	ts := httptest.NewServer(http.HandlerFunc(handlerFunc))
	sc := &SearchClient{URL: ts.URL, AccessToken: token}

	r := &SearchRequest{
		Limit:      limit,
		Offset:     offset,
		Query:      query,
		OrderField: orderField,
		OrderBy:    orderBy,
	}

	result, err := sc.FindUsers(*r)
	if err != nil {
		return nil, err
	}

	return result, nil
}
