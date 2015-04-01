package formdata

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"testing"
	"time"
)

type Object struct {
	Number int                   `formdata:"number"`
	Float  float64               `formdata:"float"`
	Ignore string                `formdata:"-"`
	String string                `formdata:"foo"`
	File   *multipart.FileHeader `formdata:"image"`
	Date   time.Time             `formdata:"date"`
	Test   *TestFormData         `formdata:"test"`
	M      map[string]string     `formdata:"m"`
	Array  []string              `formdata:"array"`
}

type TestFormData struct {
	Test int
}

func (o *TestFormData) UnmarshalFormData(value string) error {
	o.Test = 9000
	return nil
}

func TestFormdata(t *testing.T) {
	test := &Object{}
	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	number, err := w.CreateFormField("number")
	if err != nil {
		t.Error(err)
	}
	number.Write([]byte("42"))
	float, err := w.CreateFormField("float")
	if err != nil {
		t.Error(err)
	}
	float.Write([]byte("42.42"))
	m, err := w.CreateFormField("m")
	if err != nil {
		t.Error(err)
	}
	m.Write([]byte("{\"fr\":\"asdsad\"}"))
	ignore, err := w.CreateFormField("ignore")
	if err != nil {
		t.Error(err)
	}
	ignore.Write([]byte("i must be hidden"))
	foo, err := w.CreateFormField("foo")
	if err != nil {
		t.Error(err)
	}
	foo.Write([]byte(""))
	fileWritter, err := w.CreateFormFile("image", "../src/formdata/README.md")
	if err != nil {
		t.Error(err)
	}
	file, err := os.Open("./README.md")
	if err != nil {
		t.Error(err)
	}
	_, err = io.Copy(fileWritter, file)
	if err != nil {
		t.Error(err)
	}
	date, err := w.CreateFormField("date")
	if err != nil {
		t.Error(err)
	}
	date.Write([]byte("2014-09-03T14:07:59.773Z"))
	te, err := w.CreateFormField("test")
	if err != nil {
		t.Error(err)
	}
	te.Write([]byte("whatever"))
	array, err := w.CreateFormField("array")
	if err != nil {
		t.Error(err)
	}
	array.Write([]byte(`["caca", "popo"]`))
	w.Close()
	r, err := http.NewRequest("POST", "test.com", buf)
	if err != nil {
		t.Error(err)
	}
	r.Header.Set("Content-Type", w.FormDataContentType())
	if err := Unmarshal(r, test); err != nil {
		t.Error(err)
	}
	t.Logf("%#v\n", test)
}
