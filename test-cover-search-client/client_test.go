package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	_ "reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

type TestCase struct{
	ID 		string
	Result 	*SearchResponse
	Request SearchRequest
	AccessToken string
	IsError bool
}


type XMLData struct{
	XMLPersons []XMLPerson `xml:"row"`
}


type XMLPerson struct {
	Id     		int 	`xml:"id"json:"id"`
	FirstName   string 	`xml:"first_name"json:"-"`
	LastName 	string 	`xml:"last_name"json:"-"`
	Age    		int 	`xml:"age" json:"age"`
	About  		string 	`xml:"about"json:"about"`
	Gender 		string 	`xml:"gender"json:"gender"`
	Name       	string 	`xml:"-"json:"name"`
}

type RequestFilter struct{
	Limit      int 		`json:"limit,string,omitempty"`
	Offset     int  	`json:"offset,string,omitempty"`  // Можно учесть после сортировки
	Query      string 	`json:"query,omitempty"`// подстрока в 1 из полей
	OrderField string	`json:"order_field,omitempty"`
	// -1 по убыванию, 0 как встретилось, 1 по возрастанию
	OrderBy int 		`json:"order_by,string,omitempty"`
}

var secretHeader string  = "secret"

var availableOrder struct {
	Id   int
	Age  int
	Name string
}
// код писать тут
func parseXmlData(xmldata *XMLData )  {
	xmlFile, err := os.Open("dataset.xml")
	if err != nil {
		fmt.Errorf("Wrong Dataset")
	}

	defer xmlFile.Close()
	byteValue, _ := ioutil.ReadAll(xmlFile)

	xml.Unmarshal(byteValue, &xmldata)
}

func parseForm(values url.Values ) ([]byte, error) {
	requestMap := map[string]string{}
	for key, value := range values {
		requestMap[key] = value[0]
	}

	jsonData, err := json.Marshal(requestMap)
	if err != nil {
		return nil, fmt.Errorf("Request parsing error")
	}
	return jsonData, nil
}

func getAvailableOrderField() map[string]struct{} {
	aof := make(map[string]struct{})
	aof["Id"] = struct{}{}
	aof["Age"]= struct{}{}
	aof["Name"] = struct{}{}
	return aof
}

func SearchServer(w http.ResponseWriter, r *http.Request){
	availableOrder := getAvailableOrderField()

	var usersData XMLData
	parseXmlData(&usersData)
	if r.Header.Get("AccessToken") != secretHeader {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm();
	err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	formJson, err := parseForm(r.Form)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	var filters RequestFilter
	if err := json.Unmarshal(formJson, &filters);
	err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if filters.OrderBy > 1 || filters.OrderBy < -1 {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{"Error": "Wrong Order Field"}`)
		return
	}

	if _, validOrderField := availableOrder[filters.OrderField]; !validOrderField && filters.OrderField != "" {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{"Error": "ErrorBadOrderField"}`)
		return
	}

	var filteringData []XMLPerson
	for _, v :=  range usersData.XMLPersons {
		v.Name = v.FirstName + " " + v.LastName
		if strings.Contains(v.Name, filters.Query) || strings.Contains(v.About, filters.Query) {
			filteringData = append(filteringData, v)
		}

	}
	sortFilteringData(filteringData, filters.OrderField, filters.OrderBy)

	//Write response body for success response
	if len(filteringData) > filters.Limit {
		filteringData  = filteringData[0 : filters.Limit]
	}

	jsonResponse, err := json.Marshal(filteringData)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(jsonResponse)
}

func sortFilteringData(filteringData []XMLPerson, orderField string, orderBy int){
	sort.SliceStable(filteringData, func(i, j int) bool {
		if  orderField ==  "Age" && filteringData[i].Age == filteringData[j].Age {
			return  filteringData[i].Name > filteringData[j].Name
		}
		if orderBy == -1 {
			switch orderField {
			case "Age" :
				return filteringData[i].Age > filteringData[j].Age
			case "Id":
				return filteringData[i].Id > filteringData[j].Id
			case "Name":
				return filteringData[i].Name > filteringData[j].Name
			}
		}

		if orderBy == 1 {
			switch orderField {
			case "Age" :
				return filteringData[i].Age < filteringData[j].Age
			case "Id":
				return filteringData[i].Id < filteringData[j].Id
			case "Name":
				return filteringData[i].Name < filteringData[j].Name
			}
		}
		return false
	})
}

func TestFindUserMainCase(t *testing.T){
	cases := []TestCase{
		TestCase{
			ID: "__wrong_header",
			Result: nil,
			AccessToken: "wrong header",
			Request: SearchRequest{},
			IsError: true,
		},
		TestCase{
			ID: "__wrong_order_field",
			Result: nil,
			AccessToken: secretHeader,
			Request: SearchRequest{
				OrderField: "wrong",
			},
			IsError: true,
		},
		TestCase{
			ID: "__wrong_error_json",
			Result: nil,
			AccessToken: secretHeader,
			Request: SearchRequest{
				Offset: -1,
			},
			IsError: true,
		},
		TestCase{
			ID: "__wrong_request_limit",
			Result: nil,
			AccessToken: secretHeader,
			Request: SearchRequest{
				Limit: -1,
			},
			IsError: true,
		},
		TestCase{
			ID: "__wrong_order_by",
			Result: nil,
			AccessToken: secretHeader,
			Request: SearchRequest{
				OrderBy: 2,
			},
			IsError: true,
		},
		TestCase{
			ID: "__wrong_limit",
			Result: nil,
			AccessToken: secretHeader,
			Request: SearchRequest{
				Limit: 26,
			},
			IsError: true,
		},
		TestCase{
			ID: "__req_limit ",
			Result: nil,
			AccessToken: secretHeader,
			Request: SearchRequest{
				Limit: 2,
			},
			IsError: true,
		},
		TestCase{
			ID: "__req_limit ",
			Result: nil,
			AccessToken: secretHeader,
			Request: SearchRequest{
				Query: "Wolf",
				Limit: 2,
			},
			IsError: true,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	for caseNum, item := range cases{
		s := &SearchClient{
			AccessToken: item.AccessToken,
			URL:         ts.URL,
		}

		_, err := s.FindUsers(item.Request)

		if err != nil && !item.IsError {
			t.Errorf("[%d] unexpected error: %#v", caseNum, err)
		}
	}

	ts.Close()

}

func TestOrderBy(t *testing.T){
	cases := []TestCase{
		TestCase{
			ID: "__sort_query_age ",
			Result: &SearchResponse {
				Users: []User{
					{1, "Hilda Mayer", 21, "Sit commodo consectetur minim amet ex. Elit aute mollit fugiat labore sint ipsum dolor cupidatat qui reprehenderit. Eu nisi in exercitation culpa sint aliqua nulla nulla proident eu. Nisi reprehenderit anim cupidatat dolor incididunt laboris mollit magna commodo ex. Cupidatat sit id aliqua amet nisi et voluptate voluptate commodo ex eiusmod et nulla velit.\n", "female"},
					{23, "Gates Spencer", 21, "Dolore magna magna commodo irure. Proident culpa nisi veniam excepteur sunt qui et laborum tempor. Qui proident Lorem commodo dolore ipsum.\n", "male"},
				},
				NextPage: true,
			},
			AccessToken: secretHeader,
			Request: SearchRequest{
				OrderField: "Age",
				OrderBy: 1,
				Limit:2,
			},
			IsError: false,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	for caseNum, item := range cases{
		s := &SearchClient{
			AccessToken: item.AccessToken,
			URL:         ts.URL,
		}

		result, err := s.FindUsers(item.Request)

		if err != nil && !item.IsError {
			t.Errorf("[%d] unexpected error: %#v", caseNum, err)
		}
		if !reflect.DeepEqual(item.Result, result) {
			t.Errorf("[%s] wrong result, expected %#v, got %#v", item.ID, item.Result, result)
		}
	}

	ts.Close()

}

func FatalSearchServer(w http.ResponseWriter, r *http.Request){
	w.WriteHeader(http.StatusInternalServerError)
}

func unpackErrorJsonServer(w http.ResponseWriter, r *http.Request){
	w.WriteHeader(http.StatusBadRequest)
	io.WriteString(w, `{"Error": "ErrorBadOrderField}`)
}

func timeoutServerError(w http.ResponseWriter, r *http.Request){
	time.Sleep(time.Second)
}

func wrongSuccessJsonServer(w http.ResponseWriter, r *http.Request){
	io.WriteString(w, `{"Error": "ErrorBadOrderField}`)
}

func TestFatalError(t *testing.T) {
	cases := []TestCase{
		TestCase{
			ID:          "__fatal_error",
			Result:      nil,
			AccessToken: secretHeader,
			Request:     SearchRequest{},
			IsError:     true,
		},
		TestCase{
			ID:          "__unknown error",
			Result:      nil,
			AccessToken: secretHeader,
			Request:     SearchRequest{},
			IsError:     true,
		},
	}
	ts := httptest.NewServer(http.HandlerFunc(FatalSearchServer))
	var s *SearchClient
	for caseNum, item := range cases{
		if item.ID != "__unknown error"{
			s = &SearchClient{
				AccessToken: item.AccessToken,
				URL:         ts.URL,
			}
		} else {
			s = &SearchClient{
				AccessToken: item.AccessToken,
				URL:         "",
			}
		}


		_, err := s.FindUsers(item.Request)

		if err != nil && !item.IsError {
			t.Errorf("[%d] unexpected error: %#v", caseNum, err)
		}
	}

	ts.Close()
}

func TestUnpackErrorJson(t *testing.T) {
	request := SearchRequest{}
	ts := httptest.NewServer(http.HandlerFunc(unpackErrorJsonServer))

	s := &SearchClient{
		AccessToken: secretHeader,
		URL:         ts.URL,
	}

	_, err := s.FindUsers(request)

	if err == nil {
		t.Errorf("Assert UnpackErrorJson, have no error")
	}

	ts.Close()
}

func TestServerTimeout(t *testing.T) {
	request := SearchRequest{}
	ts := httptest.NewServer(http.HandlerFunc(timeoutServerError))


	s := &SearchClient{
		AccessToken: secretHeader,
		URL:         ts.URL,
	}

	_, err := s.FindUsers(request)

	if err == nil  {
		t.Errorf("Assert ServerTimeout, have no error")
	}

	ts.Close()
}

func TestWrongJsonResponse(t *testing.T) {
	request := SearchRequest{}
	ts := httptest.NewServer(http.HandlerFunc(wrongSuccessJsonServer))

	s := &SearchClient{
		AccessToken: secretHeader,
		URL:         ts.URL,
	}

	_, err := s.FindUsers(request)

	if err == nil {
		t.Errorf("Assert cant unpack result json, have no error")
	}

	ts.Close()
}