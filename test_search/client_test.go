package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

type TestCase struct {
	Request     *SearchRequest
	Result      *SearchResponse
	AccessToken string
	IsError     bool
}

type ParseUser struct {
	Id        int    `xml:"id""`
	FirstName string `xml:"first_name"`
	LastName  string `xml:"last_name"`
	Name      string `xml:"-"`
	Age       int    `xml:"age"`
	About     string `xml:"about"`
	Gender    string `xml:"gender"`
}

func (pu *ParseUser) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type Alias ParseUser
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(pu),
	}
	if err := d.DecodeElement(aux, &start); err != nil {
		return err
	}
	pu.Name = pu.FirstName + pu.LastName
	return nil
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	accessToken := r.Header.Get("AccessToken")
	if accessToken != "123" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	query := r.FormValue("query")
	orderField := r.FormValue("order_field")
	orderBy := r.FormValue("order_by")
	limit := r.FormValue("limit")
	offset := r.FormValue("offset")

	if orderField != "" && orderField != "Id" && orderField != "Age" && orderField != "Name" {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{"Error": "ErrorBadOrderField"}`)
		return
	}

	intOrderBy, err := strconv.Atoi(orderBy)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{"Error": "ErrorBadOrderBy"}`)
		return
	}

	if intOrderBy != OrderByAsc && intOrderBy != OrderByDesc && intOrderBy != OrderByAsIs {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{"Error": "ErrorBadOrderBy"}`)
		return
	}

	f, err := os.Open("./dataset.xml")
	if err != nil {
		panic(err)
	}
	decoder := xml.NewDecoder(f)
	users := []ParseUser{}
	user := ParseUser{}
	intLim, err := strconv.Atoi(limit)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	intOffset, err := strconv.Atoi(offset)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for {
		tok, tokenErr := decoder.Token()
		if tokenErr != nil && tokenErr != io.EOF {
			fmt.Println("error happend", tokenErr)
			break
		} else if tokenErr == io.EOF {
			break
		}
		if tok == nil {
			fmt.Println("t is nil break")
		}
		switch tok := tok.(type) {
		case xml.StartElement:
			if tok.Name.Local == "row" {
				if err := decoder.DecodeElement(&user, &tok); err != nil {
					fmt.Println("error happend", err)
				}

				if strings.Contains(user.Name, query) || strings.Contains(user.About, query) {
					users = append(users, user)
				}
			}
		}
	}

	if intOrderBy != OrderByAsIs {
		sort.Slice(users, func(i, j int) bool {
			switch orderField {
			case "Id":
				if intOrderBy == OrderByAsc {
					return users[i].Id < users[j].Id
				} else {
					return users[i].Id > users[j].Id
				}
			case "Age":
				if intOrderBy == OrderByAsc {
					return users[i].Age < users[j].Age
				} else {
					return users[i].Age > users[j].Age
				}
			default:
				if intOrderBy == OrderByAsc {
					return users[i].Name < users[j].Name
				} else {
					return users[i].Name > users[j].Name
				}
			}
		})
	}

	resUsers := make([]ParseUser, 0)
	offsetCounter := 0
	for _, user := range users {
		if intOffset > offsetCounter {
			offsetCounter++
			continue
		}

		if len(resUsers) >= intLim {
			break
		}

		resUsers = append(resUsers, user)
	}

	jsonUsers, err := json.Marshal(resUsers)

	w.WriteHeader(http.StatusOK)
	// w.WriteHeader(http.StatusInternalServerError)
	io.WriteString(w, string(jsonUsers))
}

func BadJsonSearchServer(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "Bad Json")
}

func BadRequestBadJsonSearchServer(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	io.WriteString(w, "Bad Json")
}

func BadSearchServer(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
}

func TimeoutSearchServer(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusGatewayTimeout)
	time.Sleep(time.Second * 6)
}

func NotExistedStatusSearchServer(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Error happend")
	w.WriteHeader(1233)
}

func TestFindUsers(t *testing.T) {
	cases := []TestCase{
		TestCase{ // Bad OrderBy
			AccessToken: "123",
			Request: &SearchRequest{
				Limit:      1,
				Offset:     0,
				Query:      "",
				OrderField: "Name",
				OrderBy:    -2,
			},
			Result:  nil,
			IsError: true,
		},
		TestCase{ // Bad Request
			AccessToken: "123",
			Request: &SearchRequest{
				Limit:      1,
				Offset:     0,
				Query:      "",
				OrderField: "test",
				OrderBy:    OrderByAsIs,
			},
			Result:  nil,
			IsError: true,
		},
		TestCase{ // Bad Access Token
			AccessToken: "321",
			Request: &SearchRequest{
				Limit:      1,
				Offset:     0,
				Query:      "",
				OrderField: "Name",
				OrderBy:    OrderByAsIs,
			},
			Result:  nil,
			IsError: true,
		},
		TestCase{ // Bad Limit
			AccessToken: "123",
			Request: &SearchRequest{
				Limit:      -1,
				Offset:     0,
				Query:      "",
				OrderField: "Name",
				OrderBy:    OrderByAsIs,
			},
			Result:  nil,
			IsError: true,
		},
		TestCase{ // Bad Offset
			AccessToken: "123",
			Request: &SearchRequest{
				Limit:      1,
				Offset:     -1,
				Query:      "",
				OrderField: "Name",
				OrderBy:    OrderByAsIs,
			},
			Result:  nil,
			IsError: true,
		},
		TestCase{ // Limit
			AccessToken: "123",
			Request: &SearchRequest{
				Limit:      26,
				Offset:     0,
				Query:      "Twil",
				OrderField: "Name",
				OrderBy:    OrderByAsIs,
			},
			Result: &SearchResponse{
				Users: []User{
					User{
						Id:     33,
						Name:   "TwilaSnow",
						Age:    36,
						About:  "Sint non sunt adipisicing sit laborum cillum magna nisi exercitation. Dolore officia esse dolore officia ea adipisicing amet ea nostrud elit cupidatat laboris. Proident culpa ullamco aute incididunt aute. Laboris et nulla incididunt consequat pariatur enim dolor incididunt adipisicing enim fugiat tempor ullamco. Amet est ullamco officia consectetur cupidatat non sunt laborum nisi in ex. Quis labore quis ipsum est nisi ex officia reprehenderit ad adipisicing fugiat. Labore fugiat ea dolore exercitation sint duis aliqua.\n",
						Gender: "female",
					},
				},
				NextPage: false,
			},
			IsError: false,
		},
		TestCase{
			AccessToken: "123",
			Request: &SearchRequest{
				Limit:      1,
				Offset:     0,
				Query:      "Twil",
				OrderField: "Name",
				OrderBy:    OrderByAsIs,
			},
			Result: &SearchResponse{
				Users: []User{
					User{
						Id:     33,
						Name:   "TwilaSnow",
						Age:    36,
						About:  "Sint non sunt adipisicing sit laborum cillum magna nisi exercitation. Dolore officia esse dolore officia ea adipisicing amet ea nostrud elit cupidatat laboris. Proident culpa ullamco aute incididunt aute. Laboris et nulla incididunt consequat pariatur enim dolor incididunt adipisicing enim fugiat tempor ullamco. Amet est ullamco officia consectetur cupidatat non sunt laborum nisi in ex. Quis labore quis ipsum est nisi ex officia reprehenderit ad adipisicing fugiat. Labore fugiat ea dolore exercitation sint duis aliqua.\n",
						Gender: "female",
					},
				},
				NextPage: false,
			},
			IsError: false,
		},
		TestCase{
			AccessToken: "123",
			Request: &SearchRequest{
				Limit:      1,
				Offset:     0,
				Query:      "en",
				OrderField: "Name",
				OrderBy:    OrderByAsIs,
			},
			Result: &SearchResponse{
				Users: []User{
					User{
						Id:     0,
						Name:   "BoydWolf",
						Age:    22,
						About:  "Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.\n",
						Gender: "male",
					},
				},
				NextPage: true,
			},
			IsError: false,
		},
		TestCase{ // Test OrderByAsIs
			AccessToken: "123",
			Request: &SearchRequest{
				Limit:      2,
				Offset:     0,
				Query:      "son",
				OrderField: "Name",
				OrderBy:    OrderByAsIs,
			},
			Result: &SearchResponse{
				Users: []User{
					User{
						Id:     10,
						Name:   "HendersonMaxwell",
						Age:    30,
						About:  "Ex et excepteur anim in eiusmod. Cupidatat sunt aliquip exercitation velit minim aliqua ad ipsum cillum dolor do sit dolore cillum. Exercitation eu in ex qui voluptate fugiat amet.\n",
						Gender: "male",
					},
					User{
						Id:     13,
						Name:   "WhitleyDavidson",
						Age:    40,
						About:  "Consectetur dolore anim veniam aliqua deserunt officia eu. Et ullamco commodo ad officia duis ex incididunt proident consequat nostrud proident quis tempor. Sunt magna ad excepteur eu sint aliqua eiusmod deserunt proident. Do labore est dolore voluptate ullamco est dolore excepteur magna duis quis. Quis laborum deserunt ipsum velit occaecat est laborum enim aute. Officia dolore sit voluptate quis mollit veniam. Laborum nisi ullamco nisi sit nulla cillum et id nisi.\n",
						Gender: "male",
					},
				},
				NextPage: true,
			},
			IsError: false,
		},
		TestCase{ // Test OrderByAsc
			AccessToken: "123",
			Request: &SearchRequest{
				Limit:      2,
				Offset:     0,
				Query:      "son",
				OrderField: "Name",
				OrderBy:    OrderByAsc,
			},
			Result: &SearchResponse{
				Users: []User{
					User{
						Id:     15,
						Name:   "AllisonValdez",
						Age:    21,
						About:  "Labore excepteur voluptate velit occaecat est nisi minim. Laborum ea et irure nostrud enim sit incididunt reprehenderit id est nostrud eu. Ullamco sint nisi voluptate cillum nostrud aliquip et minim. Enim duis esse do aute qui officia ipsum ut occaecat deserunt. Pariatur pariatur nisi do ad dolore reprehenderit et et enim esse dolor qui. Excepteur ullamco adipisicing qui adipisicing tempor minim aliquip.\n",
						Gender: "male",
					},
					User{
						Id:     30,
						Name:   "DicksonSilva",
						Age:    32,
						About:  "Ipsum aliqua proident ullamco laboris eu occaecat deserunt. Amet ut adipisicing sint veniam dolore aliquip est mollit ex officia esse eiusmod veniam. Dolore magna minim aliquip sit deserunt. Nostrud occaecat dolore aliqua aliquip voluptate aliquip ad adipisicing.\n",
						Gender: "male",
					},
				},
				NextPage: true,
			},
			IsError: false,
		},
		TestCase{ // Test OrderByDesc
			AccessToken: "123",
			Request: &SearchRequest{
				Limit:      2,
				Offset:     0,
				Query:      "son",
				OrderField: "Name",
				OrderBy:    OrderByDesc,
			},
			Result: &SearchResponse{
				Users: []User{
					User{
						Id:     13,
						Name:   "WhitleyDavidson",
						Age:    40,
						About:  "Consectetur dolore anim veniam aliqua deserunt officia eu. Et ullamco commodo ad officia duis ex incididunt proident consequat nostrud proident quis tempor. Sunt magna ad excepteur eu sint aliqua eiusmod deserunt proident. Do labore est dolore voluptate ullamco est dolore excepteur magna duis quis. Quis laborum deserunt ipsum velit occaecat est laborum enim aute. Officia dolore sit voluptate quis mollit veniam. Laborum nisi ullamco nisi sit nulla cillum et id nisi.\n",
						Gender: "male",
					},
					User{
						Id:     14,
						Name:   "NicholsonNewman",
						Age:    23,
						About:  "Tempor minim reprehenderit dolore et ad. Irure id fugiat incididunt do amet veniam ex consequat. Quis ad ipsum excepteur eiusmod mollit nulla amet velit quis duis ut irure.\n",
						Gender: "male",
					},
				},
				NextPage: true,
			},
			IsError: false,
		},
		TestCase{ // Test OrderBy Id
			AccessToken: "123",
			Request: &SearchRequest{
				Limit:      2,
				Offset:     0,
				Query:      "son",
				OrderField: "Id",
				OrderBy:    OrderByAsc,
			},
			Result: &SearchResponse{
				Users: []User{
					User{
						Id:     10,
						Name:   "HendersonMaxwell",
						Age:    30,
						About:  "Ex et excepteur anim in eiusmod. Cupidatat sunt aliquip exercitation velit minim aliqua ad ipsum cillum dolor do sit dolore cillum. Exercitation eu in ex qui voluptate fugiat amet.\n",
						Gender: "male",
					},
					User{
						Id:     13,
						Name:   "WhitleyDavidson",
						Age:    40,
						About:  "Consectetur dolore anim veniam aliqua deserunt officia eu. Et ullamco commodo ad officia duis ex incididunt proident consequat nostrud proident quis tempor. Sunt magna ad excepteur eu sint aliqua eiusmod deserunt proident. Do labore est dolore voluptate ullamco est dolore excepteur magna duis quis. Quis laborum deserunt ipsum velit occaecat est laborum enim aute. Officia dolore sit voluptate quis mollit veniam. Laborum nisi ullamco nisi sit nulla cillum et id nisi.\n",
						Gender: "male",
					},
				},
				NextPage: true,
			},
			IsError: false,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	for caseNum, item := range cases {
		c := &SearchClient{
			AccessToken: item.AccessToken,
			URL:         ts.URL,
		}
		result, err := c.FindUsers(*item.Request)
		if err != nil && !item.IsError {
			t.Errorf("[%d] unexpected error: %#v", caseNum, err)
		}
		if !reflect.DeepEqual(item.Result, result) {
			t.Errorf("[%d] wrong result, expected %#v, got %#v", caseNum, item.Result, result)
		}
	}

	ts.Close()

	handlers := []http.HandlerFunc{
		BadJsonSearchServer,
		BadRequestBadJsonSearchServer,
		BadSearchServer,
		TimeoutSearchServer,
		NotExistedStatusSearchServer,
	}

	item := TestCase{
		AccessToken: "123",
		Request: &SearchRequest{
			Limit:      1,
			Offset:     0,
			Query:      "",
			OrderField: "name",
			OrderBy:    OrderByAsc,
		},
		Result:  nil,
		IsError: true,
	}

	for i, handler := range handlers {
		ts = httptest.NewServer(http.HandlerFunc(handler))
		caseNum := len(cases) + i
		c := &SearchClient{
			AccessToken: item.AccessToken,
			URL:         ts.URL,
		}
		result, err := c.FindUsers(*item.Request)
		if err != nil && !item.IsError {
			t.Errorf("[%d] unexpected error: %#v", caseNum, err)
		}
		if !reflect.DeepEqual(item.Result, result) {
			t.Errorf("[%d] wrong result, expected %#v, got %#v", caseNum, item.Result, result)
		}
		ts.Close()
	}
}
