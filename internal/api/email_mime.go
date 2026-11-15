package api

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"mime/multipart"
	"mime/quotedprintable"
	"net/textproto"
	"strings"
)

// multipart/alternative remains available to text-only readers. When present,
// the PNG is a related inline part, referenced only by its Content-ID.
func emailMIME(bodyHTML, bodyPlain string, logo []byte) (string, []byte, error) {
	var alternative bytes.Buffer
	alt := multipart.NewWriter(&alternative)
	for _, part := range []struct{ kind, body string }{{"text/plain", bodyPlain}, {"text/html", bodyHTML}} {
		if part.kind == "text/html" && part.body == "" {
			continue
		}
		headers := textproto.MIMEHeader{"Content-Type": {part.kind + "; charset=UTF-8"}, "Content-Transfer-Encoding": {"quoted-printable"}}
		writer, err := alt.CreatePart(headers)
		if err != nil {
			return "", nil, err
		}
		encoded := quotedprintable.NewWriter(writer)
		if _, err = io.WriteString(encoded, part.body); err != nil {
			return "", nil, err
		}
		if err = encoded.Close(); err != nil {
			return "", nil, err
		}
	}
	if err := alt.Close(); err != nil {
		return "", nil, err
	}
	contentType := fmt.Sprintf("multipart/alternative; boundary=%q", alt.Boundary())
	if len(logo) == 0 {
		return contentType, alternative.Bytes(), nil
	}
	var related bytes.Buffer
	outer := multipart.NewWriter(&related)
	part, err := outer.CreatePart(textproto.MIMEHeader{"Content-Type": {contentType}})
	if err != nil {
		return "", nil, err
	}
	if _, err = part.Write(alternative.Bytes()); err != nil {
		return "", nil, err
	}
	part, err = outer.CreatePart(textproto.MIMEHeader{
		"Content-Type":              {"image/png; name=logo.png"},
		"Content-Disposition":       {"inline; filename=logo.png"},
		"Content-ID":                {"<" + emailLogoCID + ">"},
		"Content-Transfer-Encoding": {"base64"},
	})
	if err != nil {
		return "", nil, err
	}
	encoded := base64.StdEncoding.EncodeToString(logo)
	var lines strings.Builder
	for len(encoded) > 0 {
		n := min(76, len(encoded))
		lines.WriteString(encoded[:n])
		lines.WriteString("\r\n")
		encoded = encoded[n:]
	}
	if _, err = io.WriteString(part, lines.String()); err != nil {
		return "", nil, err
	}
	if err = outer.Close(); err != nil {
		return "", nil, err
	}
	return fmt.Sprintf("multipart/related; boundary=%q", outer.Boundary()), related.Bytes(), nil
}
