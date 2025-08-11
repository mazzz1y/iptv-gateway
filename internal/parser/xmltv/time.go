package xmltv

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"
)

type Time struct {
	Time time.Time
}

func (t *Time) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return xml.Attr{
		Name:  name,
		Value: t.Time.Format("20060102150405 -0700"),
	}, nil
}

func (t *Time) UnmarshalXMLAttr(attr xml.Attr) error {
	if strings.HasPrefix(attr.Value, "-") {
		return nil
	}

	t1, err := time.Parse("20060102150405 -0700", attr.Value)
	if err == nil {
		t.Time = t1
		return nil
	}

	t1, err = time.Parse("20060102150405", attr.Value)
	if err == nil {
		t.Time = t1
		return nil
	}

	t1, err = time.Parse("20060102150405 +0000", attr.Value)
	if err != nil {
		return err
	}

	t.Time = t1
	return nil
}

type Date time.Time

func (p *Date) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	t := time.Time(*p)
	if t.IsZero() {
		return nil
	}
	return e.EncodeElement(t.Format("20060102"), start)
}

func (p *Date) UnmarshalXML(d *xml.Decoder, start xml.StartElement) (err error) {
	var content string
	if e := d.DecodeElement(&content, &start); e != nil {
		return fmt.Errorf("get the type Date field of %s error", start.Name.Local)
	}

	dateFormat := "20060102"

	if len(content) == 4 {
		dateFormat = "2006"
	}

	if strings.Contains(content, "|") {
		content = strings.Split(content, "|")[0]
		dateFormat = "2006"
	}

	if v, e := time.Parse(dateFormat, content); e != nil {
		return fmt.Errorf("the type Date field of %s is not a time, value is: %s", start.Name.Local, content)
	} else {
		*p = Date(v)
	}
	return nil
}

func (p *Date) MarshalJSON() ([]byte, error) {
	t := time.Time(*p)
	str := "\"" + t.Format("20060102") + "\""
	return []byte(str), nil
}

func (p *Date) UnmarshalJSON(text []byte) (err error) {
	strDate := string(text[1 : 8+1])

	if v, e := time.Parse("20060102", strDate); e != nil {
		return fmt.Errorf("date should be a time, error value is: %s", strDate)
	} else {
		*p = Date(v)
	}
	return nil
}

func (e *ElementPresent) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	if e.Present {
		return enc.EncodeElement("", start)
	}
	return nil
}

func (e *ElementPresent) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	e.Present = true
	var content string
	return d.DecodeElement(&content, &start)
}

func (e *ElementPresent) MarshalJSON() ([]byte, error) {
	if e.Present {
		return []byte("true"), nil
	}
	return []byte("false"), nil
}

func (e *ElementPresent) UnmarshalJSON(data []byte) error {
	s := string(data)
	e.Present = s == "true"
	return nil
}
