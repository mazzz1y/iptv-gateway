package xmltv

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
)

const xmlEncoderBufferSize = 64 * 1024

type XMLEncoder struct {
	writer        *bufio.Writer
	encoder       *xml.Encoder
	headerWritten bool
	footerWritten bool
}

func NewEncoder(w io.Writer) *XMLEncoder {
	bufferedWriter := bufio.NewWriterSize(w, xmlEncoderBufferSize)
	return &XMLEncoder{
		writer:        bufferedWriter,
		encoder:       xml.NewEncoder(bufferedWriter),
		headerWritten: false,
		footerWritten: false,
	}
}

func (e *XMLEncoder) writeHeader() error {
	if e.headerWritten {
		return nil
	}

	tokens := []xml.Token{
		xml.ProcInst{Target: "xml", Inst: []byte(`version="1.0" encoding="UTF-8"`)},
		xml.Directive(`DOCTYPE tv SYSTEM "xmltv.dtd"`),
		xml.StartElement{Name: xml.Name{Local: "tv"}},
	}

	for _, token := range tokens {
		if err := e.encoder.EncodeToken(token); err != nil {
			return err
		}
	}

	e.headerWritten = true
	return nil
}

func (e *XMLEncoder) WriteFooter() error {
	if !e.headerWritten {
		return fmt.Errorf("header not written")
	}

	if e.footerWritten {
		return fmt.Errorf("footer already written")
	}

	if err := e.encoder.EncodeToken(xml.EndElement{Name: xml.Name{Local: "tv"}}); err != nil {
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
		if err := e.encoder.EncodeElement(v, xml.StartElement{Name: xml.Name{Local: "channel"}}); err != nil {
			return fmt.Errorf("encode channel: %w", err)
		}
	case Programme:
		if err := e.encoder.EncodeElement(v, xml.StartElement{Name: xml.Name{Local: "programme"}}); err != nil {
			return fmt.Errorf("encode programme: %w", err)
		}
	default:
		return fmt.Errorf("unsupported type: %T", item)
	}

	return nil
}

func (e *XMLEncoder) Close() error {
	if err := e.encoder.Flush(); err != nil {
		return err
	}

	if err := e.writer.Flush(); err != nil {
		return err
	}

	return nil
}
