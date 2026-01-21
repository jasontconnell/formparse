package formparse

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type Test struct {
	ID    string   `cookie:"test_id"`
	CID   int      `cookie:"test_id_int"`
	Query string   `query:"query"`
	Name  string   `form:"name"`
	Age   int      `form:"age"`
	Vals  []string `form:"vals"`
	Check bool     `form:"checkbox"`
	Ints  []int    `form:"ivals"`
}

func TestParse1(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := ParseForm[Test](r)

		w.WriteHeader(http.StatusOK)
		enc := json.NewEncoder(w)
		enc.Encode(t)
	}))
	defer server.Close()

	testbody := `name=Jason%20Connell&age=46&checkbox=on&vals=t1&vals=t2&vals=t3&ivals=1&ivals=333&ivals=25`

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
