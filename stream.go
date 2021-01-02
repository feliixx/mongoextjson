// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mongoextjson

import (
	"bytes"
	"io"
)

// A Decoder reads and decodes JSON values from an input stream.
type Decoder struct {
	r     io.Reader
	buf   []byte
	d     decodeState
	scanp int // start of unread data in buf
	scan  scanner
	err   error

	tokenState int
}

// NewDecoder returns a new decoder that reads from r.
//
// The decoder introduces its own buffering and may
// read data from r beyond the JSON values requested.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

// Decode reads the next JSON-encoded value from its
// input and stores it in the value pointed to by v.
//
// See the documentation for Unmarshal for details about
// the conversion of JSON into a Go value.
func (dec *Decoder) Decode(v interface{}) error {
	if dec.err != nil {
		return dec.err
	}

	if err := dec.tokenPrepareForDecode(); err != nil {
		return err
	}

	if !dec.tokenValueAllowed() {
		return &SyntaxError{msg: "not at beginning of value"}
	}

	// Read whole value into buffer.
	n, err := dec.readValue()
	if err != nil {
		return err
	}
	dec.d.init(dec.buf[dec.scanp : dec.scanp+n])
	dec.scanp += n

	// Don't save err from unmarshal into dec.err:
	// the connection is still usable since we read a complete JSON
	// object from it before the error happened.
	err = dec.d.unmarshal(v)

	// fixup token streaming state
	dec.tokenValueEnd()

	return err
}

// Buffered returns a reader of the data remaining in the Decoder's
// buffer. The reader is valid until the next call to Decode.
func (dec *Decoder) Buffered() io.Reader {
	return bytes.NewReader(dec.buf[dec.scanp:])
}

// readValue reads a JSON value into dec.buf.
// It returns the length of the encoding.
func (dec *Decoder) readValue() (int, error) {
	dec.scan.reset()

	scanp := dec.scanp
	var err error
Input:
	for {
		// Look in the buffer for a new value.
		for i, c := range dec.buf[scanp:] {
			dec.scan.bytes++
			v := dec.scan.step(&dec.scan, c)
			if v == scanEnd {
				scanp += i
				break Input
			}
			// scanEnd is delayed one byte.
			// We might block trying to get that byte from src,
			// so instead invent a space byte.
			if (v == scanEndObject || v == scanEndArray) && dec.scan.step(&dec.scan, ' ') == scanEnd {
				scanp += i + 1
				break Input
			}
			if v == scanError {
				dec.err = dec.scan.err
				return 0, dec.scan.err
			}
		}
		scanp = len(dec.buf)

		// Did the last read have an error?
		// Delayed until now to allow buffer scan.
		if err != nil {
			if err == io.EOF {
				if dec.scan.step(&dec.scan, ' ') == scanEnd {
					break Input
				}
				if nonSpace(dec.buf) {
					err = io.ErrUnexpectedEOF
				}
			}
			dec.err = err
			return 0, err
		}

		n := scanp - dec.scanp
		err = dec.refill()
		scanp = dec.scanp + n
	}
	return scanp - dec.scanp, nil
}

func (dec *Decoder) refill() error {
	// Make room to read more into the buffer.
	// First slide down data already consumed.
	if dec.scanp > 0 {
		n := copy(dec.buf, dec.buf[dec.scanp:])
		dec.buf = dec.buf[:n]
		dec.scanp = 0
	}

	// Grow buffer if not large enough.
	const minRead = 512
	if cap(dec.buf)-len(dec.buf) < minRead {
		newBuf := make([]byte, len(dec.buf), 2*cap(dec.buf)+minRead)
		copy(newBuf, dec.buf)
		dec.buf = newBuf
	}

	// Read. Delay error for next iteration (after scan).
	n, err := dec.r.Read(dec.buf[len(dec.buf):cap(dec.buf)])
	dec.buf = dec.buf[0 : len(dec.buf)+n]

	return err
}

func nonSpace(b []byte) bool {
	for _, c := range b {
		if !isSpace(c) {
			return true
		}
	}
	return false
}

// An Encoder writes JSON values to an output stream.
type Encoder struct {
	w          io.Writer
	err        error
	escapeHTML bool

	ext Extension
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w, escapeHTML: true}
}

// Encode writes the JSON encoding of v to the stream,
// followed by a newline character.
//
// See the documentation for Marshal for details about the
// conversion of Go values to JSON.
func (enc *Encoder) Encode(v interface{}) error {
	if enc.err != nil {
		return enc.err
	}
	e := newEncodeState()
	e.ext = enc.ext
	err := e.marshal(v, encOpts{escapeHTML: enc.escapeHTML})
	if err != nil {
		return err
	}

	// Terminate each value with a newline.
	// This makes the output look a little nicer
	// when debugging, and some kind of space
	// is required if the encoded value was a number,
	// so that the reader knows there aren't more
	// digits coming.

	// no need for this
	//e.WriteByte('\n')

	if _, err = enc.w.Write(e.Bytes()); err != nil {
		enc.err = err
	}
	encodeStatePool.Put(e)
	return err
}

// DisableHTMLEscaping causes the encoder not to escape angle brackets
// ("<" and ">") or ampersands ("&") in JSON strings.
func (enc *Encoder) DisableHTMLEscaping() {
	enc.escapeHTML = false
}

// A Token holds a value of one of these types:
//
//	Delim, for the four JSON delimiters [ ] { }
//	bool, for JSON booleans
//	float64, for JSON numbers
//	Number, for JSON numbers
//	string, for JSON string literals
//	nil, for JSON null
//
type Token interface{}

const (
	tokenTopValue = iota
	tokenArrayStart
	tokenArrayValue
	tokenArrayComma
	tokenObjectColon
	tokenObjectValue
	tokenObjectComma
)

// advance tokenstate from a separator state to a value state
func (dec *Decoder) tokenPrepareForDecode() error {
	// Note: Not calling peek before switch, to avoid
	// putting peek into the standard Decode path.
	// peek is only called when using the Token API.
	switch dec.tokenState {
	case tokenArrayComma:
		c, err := dec.peek()
		if err != nil {
			return err
		}
		if c != ',' {
			return &SyntaxError{"expected comma after array element", 0}
		}
		dec.scanp++
		dec.tokenState = tokenArrayValue
	case tokenObjectColon:
		c, err := dec.peek()
		if err != nil {
			return err
		}
		if c != ':' {
			return &SyntaxError{"expected colon after object key", 0}
		}
		dec.scanp++
		dec.tokenState = tokenObjectValue
	}
	return nil
}

func (dec *Decoder) tokenValueAllowed() bool {
	switch dec.tokenState {
	case tokenTopValue, tokenArrayStart, tokenArrayValue, tokenObjectValue:
		return true
	}
	return false
}

func (dec *Decoder) tokenValueEnd() {
	switch dec.tokenState {
	case tokenArrayStart, tokenArrayValue:
		dec.tokenState = tokenArrayComma
	case tokenObjectValue:
		dec.tokenState = tokenObjectComma
	}
}

func (dec *Decoder) peek() (byte, error) {
	var err error
	for {
		for i := dec.scanp; i < len(dec.buf); i++ {
			c := dec.buf[i]
			if isSpace(c) {
				continue
			}
			dec.scanp = i
			return c, nil
		}
		// buffer has been scanned, now report any error
		if err != nil {
			return 0, err
		}
		err = dec.refill()
	}
}
