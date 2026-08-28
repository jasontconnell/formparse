package formparse

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type Embedded struct {
	EmbeddedVal string `form:"embedded"`
}

func (e *Embedded) String() string {
	return "Embedded"
}

type Test struct {
	Embedded
	ID        string         `cookie:"test_id"`
	CID       int            `cookie:"test_id_int"`
	Query     string         `query:"query"`
	Name      string         `form:"name"`
	Age       int            `form:"age"`
	Vals      []string       `form:"vals"`
	Check     bool           `form:"checkbox"`
	Ints      []int          `form:"ivals"`
	Internals []TestInternal `form:"internals"`
}

type TestInternal struct {
	String string `form:"string"`
	Int    int    `form:"int"`
}

type TestArray struct {
	Name string `form:"name"`
	Age  int    `form:"age"`
}

func TestParse(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parsed := ParseForm[Test](r)

		w.WriteHeader(http.StatusOK)
		enc := json.NewEncoder(w)
		enc.Encode(parsed)
	}))
	defer server.Close()

	testbody := `name=Jason%20Connell&age=46&checkbox=on&vals=t1&vals=t2&vals=t3&ivals=1&ivals=333&ivals=25&embedded=Embedded%20value&internals_string_0=jason&internals_int_0=4&internals_string_1=test&internals_int_1=11`

	b := bytes.NewBufferString(testbody)

	req, err := http.NewRequest("POST", server.URL+"/some/path?query=qval", b)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	c := &http.Cookie{
		Name:  "test_id",
		Value: "1234",
	}
	c2 := &http.Cookie{
		Name:  "test_id_int",
		Value: "1234",
	}

	req.AddCookie(c)
	req.AddCookie(c2)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(body))
}

func TestParseArray(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parsed := ParseFormArray[TestArray](r)

		w.WriteHeader(http.StatusOK)
		enc := json.NewEncoder(w)
		enc.Encode(parsed)
	}))
	defer server.Close()

	testbody := `age_0=46&age_1=47&age_2=48&name_0=jason&name_1=thomas&name_2=erin`

	b := bytes.NewBufferString(testbody)

	req, err := http.NewRequest("POST", server.URL+"/some/path?query=qval", b)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(body))
}
