package xmltv

import (
	"encoding/xml"
	"fmt"
	"io"
)

type Encoder interface {
	Encode(item any) error
	Close() error
}

type XMLEncoder struct {
	writer        io.Writer
	encoder       *xml.Encoder
	headerWritten bool
	footerWritten bool
}

func NewEncoder(w io.Writer) Encoder {
	return &XMLEncoder{
		writer:        w,
		encoder:       xml.NewEncoder(w),
		headerWritten: false,
		footerWritten: false,
	}
}

func (e *XMLEncoder) writeHeader() error {
	if e.headerWritten {
		return nil
	}

	_, err := io.WriteString(e.writer,
		"<?xml version=\"1.0\" encoding=\"UTF-8\"?><!DOCTYPE tv SYSTEM \"xmltv.dtd\">\n")
	if err != nil {
		return err
	}

	if err := e.encoder.EncodeToken(xml.StartElement{Name: xml.Name{Local: "tv"}}); err != nil {
		return err
	}

	e.headerWritten = true
	return nil
}

func (e *XMLEncoder) writeFooter() error {
	if e.footerWritten || !e.headerWritten {
		return nil
	}

	if err := e.encoder.EncodeToken(xml.EndElement{Name: xml.Name{Local: "tv"}}); err != nil {
		return err
	}

	if err := e.encoder.Flush(); err != nil {
		return err
	}

	e.footerWritten = true
	return nil
}

func (e *XMLEncoder) Encode(item any) error {
	if err := e.writeHeader(); err != nil {
		return err
	}

	switch v := item.(type) {
	case Channel:
		start := xml.StartElement{Name: xml.Name{Local: "channel"}}
		if err := e.encoder.EncodeElement(v, start); err != nil {
			return fmt.Errorf("error encoding channel: %w", err)
		}
		return nil
	case Programme:
		start := xml.StartElement{Name: xml.Name{Local: "programme"}}
		if err := e.encoder.EncodeElement(v, start); err != nil {
			return fmt.Errorf("error encoding programme: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("unsupported item type: %T", item)
	}
}

func (e *XMLEncoder) Close() error {
	return e.writeFooter()
}
