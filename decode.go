package formdata

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"github.com/gocarina/structs"
	"time"
)

// Unmarshaler is the interface to implements to deal with custom unmarshaling of form data.
type Unmarshaler interface {
	UnmarshalFormData(string) error
}

var (
	unmarshalerType = reflect.TypeOf(new(Unmarshaler)).Elem()
)

var (
	// TagName is the tag used to register custom names for field.
	TagName = "formdata"
)

type tag struct {
	Name   string
	IsFile bool
}

func createTagError(field *structs.Field) error {
	return fmt.Errorf("%v: too many tags", field.Name())
}

func getFieldTags(field *structs.Field) *tag {
	var name string
	var isFile bool
	tags := strings.Split(field.Tag(TagName), ",")
	switch len(tags) {
	case 0:
		name = field.Name()
	case 1:
		name = tags[0]
	default:
		panic(createTagError(field))
	}
	switch field.Value().(type) {
	case multipart.FileHeader, *multipart.FileHeader, []*multipart.FileHeader:
		isFile = true
	}
	return &tag{
		Name:   name,
		IsFile: isFile,
	}
}

// Unmarshal takes the request, parses the multipart form data and fill the object out.
// Here is a complete example:
//		type Foo struct {
//			Name string `formdata:"name"`
//			Image *multipart.FileHeader `formdata:"image"`
//		}
//
//
//		foo := &Foo{}
//		if err := formdata.Unmarshal(request, out); err != nil {
//			// deal with error
//		}
//
func Unmarshal(request *http.Request, out interface{}) error {
	if err := request.ParseMultipartForm(32 << 2); err != nil {
		return err
	}
	formValues := request.MultipartForm
	structInfo := structs.New(out)
	for _, field := range structInfo.Fields() {
		var err error
		tagInfo := getFieldTags(field)
		if tagInfo.Name == "-" || (len(formValues.Value[tagInfo.Name]) == 0 && formValues.File[tagInfo.Name] == nil) {
			continue
		}
		if tagInfo.IsFile {
			err = unmarshalFile(field, formValues.File[tagInfo.Name]...)
		} else {
			err = unmarshalText(field, formValues.Value[tagInfo.Name][0])
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func unmarshalText(field *structs.Field, textValue string) (err error) {
	fieldType := reflect.TypeOf(field.Value())
	if fieldType.Implements(unmarshalerType) {
		fieldValue := reflect.ValueOf(field.Value())
		if fieldValue.IsNil() {
			fieldValue = reflect.New(fieldValue.Type().Elem())
			if err := field.Set(fieldValue.Interface()); err != nil {
				return err
			}
		}
		err = field.Value().(Unmarshaler).UnmarshalFormData(textValue)
		return err
	}
	switch field.Kind() {
	default:
	case reflect.String:
		field.Set(textValue)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		valueInt, err := strconv.ParseInt(textValue, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(valueInt)
	case reflect.Float64, reflect.Float32:
		valueFloat, err := strconv.ParseFloat(textValue, 64)
		if err != nil {
			return err
		}
		if err := field.SetFloat(valueFloat); err != nil {
			return err
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		valueUint, err := strconv.ParseUint(textValue, 10, 64)
		if err != nil {
			return err
		}
		if err := field.SetUint(valueUint); err != nil {
			return err
		}
	case reflect.Bool:
		valueBool, err := strconv.ParseBool(textValue)
		if err != nil {
			return err
		}
		if err := field.Set(valueBool); err != nil {
			return err
		}
	case reflect.Struct:
		switch field.Value().(type) {
		case time.Time:
			date, err := time.Parse(time.RFC3339, textValue)
			if err != nil {
				return err
			}
			field.Set(date)
		}
	}
	return nil
}

func unmarshalFile(field *structs.Field, fileValues ...*multipart.FileHeader) (err error) {
	fieldValue := field.Value()
	switch fieldValue.(type) {
	case *multipart.FileHeader:
		err = field.Set(fileValues[0])
	case []*multipart.FileHeader:
		err = field.Set(fileValues)
	}
	return err
}
