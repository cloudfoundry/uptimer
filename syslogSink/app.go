package syslogSink

const Source = `
package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"time"
	"unicode"
	"unicode/utf8"
)

func main() {
	l, err := net.Listen("tcp4", fmt.Sprintf(":%s", os.Getenv("PORT")))
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			continue
		}

		go func() {
			defer conn.Close()
			handleConnection(conn)
		}()
	}
}

func handleConnection(conn net.Conn) {
	msg := Message{}
	for {
		if _, err := msg.ReadFrom(conn); err != nil {
			return
		}

		fmt.Println(string(msg.Message))
	}
}

// The following code is sourced from code.cloudfoundry.org/rfc5424
// and is copied here to minimize dependencies in apps we push:

// allowLongSdNames is true to allow names longer than the RFC-specified limit
// of 32-characters. (When true, this violates RFC-5424).
const allowLongSdNames = true

// RFC5424TimeOffsetNum is the timestamp defined by RFC-5424 with the
// NUMOFFSET instead of Z.
const RFC5424TimeOffsetNum = "2006-01-02T15:04:05.999999-07:00"

// RFC5424TimeOffsetUTC is the timestamp defined by RFC-5424 with the offset
// set to 0 for UTC.
const RFC5424TimeOffsetUTC = "2006-01-02T15:04:05.999999Z"

// ErrInvalidValue is returned when a log message cannot be emitted because one
// of the values is invalid.
type ErrInvalidValue struct {
	Property string
	Value    interface{}
}

func (e ErrInvalidValue) Error() string {
	return fmt.Sprintf("Message cannot be serialized because %s is invalid: %v",
		e.Property, e.Value)
}

// invalidValue returns an invalid value error with the given property
func invalidValue(property string, value interface{}) error {
	return ErrInvalidValue{Property: property, Value: value}
}

func nilify(x string) string {
	if x == "" {
		return "-"
	}
	return x
}

func escapeSDParam(s string) string {
	escapeCount := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\\', '"', ']':
			escapeCount++
		}
	}
	if escapeCount == 0 {
		return s
	}

	t := make([]byte, len(s)+escapeCount)
	j := 0
	for i := 0; i < len(s); i++ {
		switch c := s[i]; c {
		case '\\', '"', ']':
			t[j] = '\\'
			t[j+1] = c
			j += 2
		default:
			t[j] = s[i]
			j++
		}
	}
	return string(t)
}

func isPrintableUsASCII(s string) bool {
	for _, ch := range s {
		if ch < 33 || ch > 126 {
			return false
		}
	}
	return true
}

func isValidSdName(s string) bool {
	if !allowLongSdNames && len(s) > 32 {
		return false
	}
	for _, ch := range s {
		if ch < 33 || ch > 126 {
			return false
		}
		if ch == '=' || ch == ']' || ch == '"' {
			return false
		}
	}
	return true
}

func (m Message) assertValid() error {

	// HOSTNAME        = NILVALUE / 1*255PRINTUSASCII
	if !isPrintableUsASCII(m.Hostname) {
		return invalidValue("Hostname", m.Hostname)
	}
	if len(m.Hostname) > 255 {
		return invalidValue("Hostname", m.Hostname)
	}

	// APP-NAME        = NILVALUE / 1*48PRINTUSASCII
	if !isPrintableUsASCII(m.AppName) {
		return invalidValue("AppName", m.AppName)
	}
	if len(m.AppName) > 48 {
		return invalidValue("AppName", m.AppName)
	}

	// PROCID          = NILVALUE / 1*128PRINTUSASCII
	if !isPrintableUsASCII(m.ProcessID) {
		return invalidValue("ProcessID", m.ProcessID)
	}
	if len(m.ProcessID) > 128 {
		return invalidValue("ProcessID", m.ProcessID)
	}

	// MSGID           = NILVALUE / 1*32PRINTUSASCII
	if !isPrintableUsASCII(m.MessageID) {
		return invalidValue("MessageID", m.MessageID)
	}
	if len(m.MessageID) > 32 {
		return invalidValue("MessageID", m.MessageID)
	}

	for _, sdElement := range m.StructuredData {
		if !isValidSdName(sdElement.ID) {
			return invalidValue("StructuredData/ID", sdElement.ID)
		}
		for _, sdParam := range sdElement.Parameters {
			if !isValidSdName(sdParam.Name) {
				return invalidValue("StructuredData/Name", sdParam.Name)
			}
			if !utf8.ValidString(sdParam.Value) {
				return invalidValue("StructuredData/Value", sdParam.Value)
			}
		}
	}
	return nil
}

// MarshalBinary marshals the message to a byte slice, or returns an error
func (m Message) MarshalBinary() ([]byte, error) {
	if err := m.assertValid(); err != nil {
		return nil, err
	}

	b := bytes.NewBuffer(nil)
	fmt.Fprintf(b, "<%d>1 %s %s %s %s %s ",
		m.Priority,
		m.Timestamp.Format(RFC5424TimeOffsetNum),
		nilify(m.Hostname),
		nilify(m.AppName),
		nilify(m.ProcessID),
		nilify(m.MessageID))

	if len(m.StructuredData) == 0 {
		fmt.Fprint(b, "-")
	}
	for _, sdElement := range m.StructuredData {
		fmt.Fprintf(b, "[%s", sdElement.ID)
		for _, sdParam := range sdElement.Parameters {
			fmt.Fprintf(b, " %s=\"%s\"", sdParam.Name,
				escapeSDParam(sdParam.Value))
		}
		fmt.Fprintf(b, "]")
	}

	if len(m.Message) > 0 {
		fmt.Fprint(b, " ")
		b.Write(m.Message)
	}
	return b.Bytes(), nil
}

// Pacakge rfc5424 is a library for parsing and serializing RFC-5424 structured
// syslog messages.
//
// Example usage:
//
//     m := rfc5424.Message{
//         Priority:  rfc5424.Daemon | rfc5424.Info,
//         Timestamp: time.Now(),
//         Hostname:  "myhostname",
//         AppName:   "someapp",
//         Message:   []byte("Hello, World!"),
//     }
//     m.AddDatum("foo@1234", "Revision", "1.2.3.4")
//     m.WriteTo(os.Stdout)
//
// Produces output like:
//
//     107 <7>1 2016-02-28T09:57:10.804642398-05:00 myhostname someapp - - [foo@1234 Revision="1.2.3.4"] Hello, World!
//
// You can also use the library to parse syslog messages:
//
//     m := rfc5424.Message{}
//     _, err := m.ReadFrom(os.Stdin)
//     fmt.Printf("%s\n", m.Message)

// Message represents a log message as defined by RFC-5424
// (https://tools.ietf.org/html/rfc5424)
type Message struct {
	Priority       Priority
	Timestamp      time.Time
	Hostname       string
	AppName        string
	ProcessID      string
	MessageID      string
	StructuredData []StructuredData
	Message        []byte
}

// SDParam represents parameters for structured data
type SDParam struct {
	Name  string
	Value string
}

// StructuredData represents structured data within a log message
type StructuredData struct {
	ID         string
	Parameters []SDParam
}

// AddDatum adds structured data to a log message
func (m *Message) AddDatum(ID string, Name string, Value string) {
	if m.StructuredData == nil {
		m.StructuredData = []StructuredData{}
	}
	for i, sd := range m.StructuredData {
		if sd.ID == ID {
			sd.Parameters = append(sd.Parameters, SDParam{Name: Name, Value: Value})
			m.StructuredData[i] = sd
			return
		}
	}

	m.StructuredData = append(m.StructuredData, StructuredData{
		ID: ID,
		Parameters: []SDParam{
			{
				Name:  Name,
				Value: Value,
			},
		},
	})
}

// WriteTo writes the message to a stream of messages in the style defined
// by RFC-5425. (It does not implement the TLS stuff described in the RFC, just
// the length delimiting.
func (m Message) WriteTo(w io.Writer) (int64, error) {
	b, err := m.MarshalBinary()
	if err != nil {
		return 0, err
	}
	n, err := fmt.Fprintf(w, "%d %s", len(b), b)

	return int64(n), err
}

func readUntilSpace(r io.Reader) ([]byte, int, error) {
	buf := []byte{}
	nbytes := 0
	for {
		b := []byte{0}
		n, err := r.Read(b)
		nbytes += n
		if err != nil {
			return nil, nbytes, err
		}
		if b[0] == ' ' {
			return buf, nbytes, nil
		}
		buf = append(buf, b...)
	}
}

// ReadFrom reads a single record from an RFC-5425 style stream of messages
func (m *Message) ReadFrom(r io.Reader) (int64, error) {
	lengthBuf, n1, err := readUntilSpace(r)
	if err != nil {
		return 0, err
	}
	length, err := strconv.Atoi(string(lengthBuf))
	if err != nil {
		return 0, err
	}
	r2 := io.LimitReader(r, int64(length))
	buf, err := io.ReadAll(r2)
	if err != nil {
		return int64(n1 + len(buf)), err
	}
	if len(buf) != int(length) {
		return int64(n1 + len(buf)), fmt.Errorf("Expected to read %d bytes, got %d", length, len(buf))
	}
	err = m.UnmarshalBinary(buf)
	if err != nil {
		return 0, err
	}
	return int64(n1 + len(buf)), err
}

// ErrBadFormat is the error that is returned when a log message cannot be parsed
type ErrBadFormat struct {
	Property string
}

func (e ErrBadFormat) Error() string {
	return fmt.Sprintf("Message cannot be unmarshaled because it is not well formed (%s)",
		e.Property)
}

// badFormat returns a bad format error with the given property
func badFormat(property string) error {
	return ErrBadFormat{Property: property}
}

// UnmarshalBinary unmarshals a byte slice into a message
func (m *Message) UnmarshalBinary(inputBuffer []byte) error {
	r := bytes.NewBuffer(inputBuffer)

	// RFC-5424
	// SYSLOG-MSG      = HEADER SP STRUCTURED-DATA [SP MSG]
	if err := m.readHeader(r); err != nil {
		return err
	}

	if err := readSpace(r); err != nil {
		return err // unreachable
	}
	if err := m.readStructuredData(r); err != nil {
		return err
	}

	// MSG is optional
	ch, _, err := r.ReadRune()
	if err == io.EOF {
		return nil
	} else if ch != ' ' {
		return badFormat("MSG") // unreachable
	}

	// TODO(ross): detect and handle UTF-8 BOM (\xef\xbb\xbf)
	//
	// MSG             = MSG-ANY / MSG-UTF8
	// MSG-ANY         = *OCTET ; not starting with BOM
	// MSG-UTF8        = BOM UTF-8-STRING
	// BOM             = %xEF.BB.BF

	// To be on the safe side, remaining stuff is copied over
	m.Message = copyFrom(r.Bytes())
	return nil
}

// readHeader reads a HEADER as defined in RFC-5424
//
// HEADER          = PRI VERSION SP TIMESTAMP SP HOSTNAME
// SP APP-NAME SP PROCID SP MSGID
// PRI             = "<" PRIVAL ">"
// PRIVAL          = 1*3DIGIT ; range 0 .. 191
// VERSION         = NONZERO-DIGIT 0*2DIGIT
// HOSTNAME        = NILVALUE / 1*255PRINTUSASCII
//
// APP-NAME        = NILVALUE / 1*48PRINTUSASCII
// PROCID          = NILVALUE / 1*128PRINTUSASCII
// MSGID           = NILVALUE / 1*32PRINTUSASCII
//
// TIMESTAMP       = NILVALUE / FULL-DATE "T" FULL-TIME
// FULL-DATE       = DATE-FULLYEAR "-" DATE-MONTH "-" DATE-MDAY
// DATE-FULLYEAR   = 4DIGIT
// DATE-MONTH      = 2DIGIT  ; 01-12
// DATE-MDAY       = 2DIGIT  ; 01-28, 01-29, 01-30, 01-31 based on
// ; month/year
// FULL-TIME       = PARTIAL-TIME TIME-OFFSET
// PARTIAL-TIME    = TIME-HOUR ":" TIME-MINUTE ":" TIME-SECOND
// [TIME-SECFRAC]
// TIME-HOUR       = 2DIGIT  ; 00-23
// TIME-MINUTE     = 2DIGIT  ; 00-59
// TIME-SECOND     = 2DIGIT  ; 00-59
// TIME-SECFRAC    = "." 1*6DIGIT
// TIME-OFFSET     = "Z" / TIME-NUMOFFSET
// TIME-NUMOFFSET  = ("+" / "-") TIME-HOUR ":" TIME-MINUTE
//
func (m *Message) readHeader(r io.RuneScanner) error {
	if err := m.readPriority(r); err != nil {
		return err
	}
	if err := m.readVersion(r); err != nil {
		return err
	}
	if err := readSpace(r); err != nil {
		return err // unreachable
	}
	if err := m.readTimestamp(r); err != nil {
		return err
	}
	if err := readSpace(r); err != nil {
		return err // unreachable
	}
	if err := m.readHostname(r); err != nil {
		return err
	}
	if err := readSpace(r); err != nil {
		return err // unreachable
	}
	if err := m.readAppName(r); err != nil {
		return err
	}
	if err := readSpace(r); err != nil {
		return err // unreachable
	}
	if err := m.readProcID(r); err != nil {
		return err
	}
	if err := readSpace(r); err != nil {
		return err // unreachable
	}
	if err := m.readMsgID(r); err != nil {
		return err
	}
	return nil
}

// readPriority reads the PRI as defined in RFC-5424 and assigns Severity and
// Facility accordingly.
func (m *Message) readPriority(r io.RuneScanner) error {
	ch, _, err := r.ReadRune()
	if err != nil {
		return err
	}
	if ch != '<' {
		return badFormat("Priority")
	}

	rv := &bytes.Buffer{}
	for {
		ch, _, err := r.ReadRune()
		if err != nil {
			return err
		}
		if unicode.IsDigit(ch) {
			rv.WriteRune(ch)
			continue
		}
		if ch != '>' {
			return badFormat("Priority")
		}

		// We have a complete integer expression
		priority, err := strconv.ParseInt(string(rv.Bytes()), 10, 32)
		if err != nil {
			return badFormat("Priority")
		}
		m.Priority = Priority(priority)
		return nil
	}
}

// readVersion reads the version string fails if it isn't 1
func (m *Message) readVersion(r io.RuneScanner) error {
	ch, _, err := r.ReadRune()
	if err != nil {
		return err
	}
	if ch != '1' {
		return badFormat("Version")
	}
	return nil
}

// readTimestamp reads a TIMESTAMP as defined in RFC-5424 and assigns
// m.Timestamp
//
// TIMESTAMP       = NILVALUE / FULL-DATE "T" FULL-TIME
// FULL-DATE       = DATE-FULLYEAR "-" DATE-MONTH "-" DATE-MDAY
// DATE-FULLYEAR   = 4DIGIT
// DATE-MONTH      = 2DIGIT  ; 01-12
// DATE-MDAY       = 2DIGIT  ; 01-28, 01-29, 01-30, 01-31 based on
//                           ; month/year
// FULL-TIME       = PARTIAL-TIME TIME-OFFSET
// PARTIAL-TIME    = TIME-HOUR ":" TIME-MINUTE ":" TIME-SECOND
// [TIME-SECFRAC]
// TIME-HOUR       = 2DIGIT  ; 00-23
// TIME-MINUTE     = 2DIGIT  ; 00-59
// TIME-SECOND     = 2DIGIT  ; 00-59
// TIME-SECFRAC    = "." 1*6DIGIT
// TIME-OFFSET     = "Z" / TIME-NUMOFFSET
// TIME-NUMOFFSET  = ("+" / "-") TIME-HOUR ":" TIME-MINUTE
func (m *Message) readTimestamp(r io.RuneScanner) error {
	timestampString, err := readWord(r)
	if err != nil {
		return err
	}

	m.Timestamp, err = time.Parse(RFC5424TimeOffsetNum, timestampString)
	if err == nil {
		return nil
	}

	m.Timestamp, err = time.Parse(RFC5424TimeOffsetUTC, timestampString)
	if err == nil {
		return nil
	}

	return err
}

func (m *Message) readHostname(r io.RuneScanner) (err error) {
	m.Hostname, err = readWord(r)
	return err
}

func (m *Message) readAppName(r io.RuneScanner) (err error) {
	m.AppName, err = readWord(r)
	return err
}

func (m *Message) readProcID(r io.RuneScanner) (err error) {
	m.ProcessID, err = readWord(r)
	return err
}

func (m *Message) readMsgID(r io.RuneScanner) (err error) {
	m.MessageID, err = readWord(r)
	return err
}

// readStructuredData reads a STRUCTURED-DATA (as defined in RFC-5424)
// from r and assigns the StructuredData member.
//
// STRUCTURED-DATA = NILVALUE / 1*SD-ELEMENT
// SD-ELEMENT      = "[" SD-ID *(SP SD-PARAM) "]"
// SD-PARAM        = PARAM-NAME "=" %d34 PARAM-VALUE %d34
// SD-ID           = SD-NAME
// PARAM-NAME      = SD-NAME
// PARAM-VALUE     = UTF-8-STRING ; characters '"', '\' and ']' MUST be escaped.
// SD-NAME         = 1*32PRINTUSASCII except '=', SP, ']', %d34 (")
func (m *Message) readStructuredData(r io.RuneScanner) (err error) {
	m.StructuredData = []StructuredData{}

	ch, _, err := r.ReadRune()
	if err != nil {
		return err
	}
	if ch == '-' {
		return nil
	}
	r.UnreadRune()

	for {
		ch, _, err := r.ReadRune()
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err // hard to reach without underlying IO error
		} else if ch == ' ' {
			r.UnreadRune()
			return nil
		} else if ch == '[' {
			r.UnreadRune()
			sde, err := readSDElement(r)
			if err != nil {
				return err
			}
			m.StructuredData = append(m.StructuredData, sde)
		} else {
			return badFormat("StructuredData")
		}
	}
}

// readSDElement reads an SD-ELEMENT as defined by RFC-5424
//
// SD-ELEMENT      = "[" SD-ID *(SP SD-PARAM) "]"
// SD-PARAM        = PARAM-NAME "=" %d34 PARAM-VALUE %d34
// SD-ID           = SD-NAME
// PARAM-NAME      = SD-NAME
// PARAM-VALUE     = UTF-8-STRING ; characters '"', '\' and ']' MUST be escaped.
// SD-NAME         = 1*32PRINTUSASCII except '=', SP, ']', %d34 (")
func readSDElement(r io.RuneScanner) (element StructuredData, err error) {
	ch, _, err := r.ReadRune()
	if err != nil {
		return element, err // hard to reach without underlying IO error
	}
	if ch != '[' {
		return element, badFormat("StructuredData[]") // unreachable
	}
	element.ID, err = readSdID(r)
	if err != nil {
		return element, err
	}
	for {
		ch, _, err := r.ReadRune()
		if err != nil {
			return element, err
		} else if ch == ']' {
			return element, nil
		} else if ch == ' ' {
			param, err := readSdParam(r)
			if err != nil {
				return element, err
			}
			element.Parameters = append(element.Parameters, *param)
		} else {
			return element, badFormat("StructuredData[]")
		}
	}
}

// readSDID reads an SD-ID as defined by RFC-5424
// SD-ID           = SD-NAME
// SD-NAME         = 1*32PRINTUSASCII except '=', SP, ']', %d34 (")
func readSdID(r io.RuneScanner) (string, error) {
	rv := &bytes.Buffer{}
	for {
		ch, _, err := r.ReadRune()
		if err != nil {
			return "", err
		}
		if ch == ' ' || ch == ']' {
			r.UnreadRune()
			return string(rv.Bytes()), nil
		}
		rv.WriteRune(ch)
	}
}

// readSdParam reads an SD-PARAM as defined by RFC-5424
// SD-PARAM        = PARAM-NAME "=" %d34 PARAM-VALUE %d34
// SD-ID           = SD-NAME
// PARAM-NAME      = SD-NAME
// PARAM-VALUE     = UTF-8-STRING ; characters '"', '\' and ']' MUST be escaped.
// SD-NAME         = 1*32PRINTUSASCII except '=', SP, ']', %d34 (")
func readSdParam(r io.RuneScanner) (sdp *SDParam, err error) {
	sdp = &SDParam{}
	sdp.Name, err = readSdParamName(r)
	if err != nil {
		return nil, err
	}
	ch, _, err := r.ReadRune()
	if err != nil {
		return nil, err // hard to reach
	}
	if ch != '=' {
		return nil, badFormat("StructuredData[].Parameters") // not reachable
	}

	sdp.Value, err = readSdParamValue(r)
	if err != nil {
		return nil, err
	}
	return sdp, nil
}

// readSdParam reads a PARAM-NAME as defined by RFC-5424
// SD-PARAM        = PARAM-NAME "=" %d34 PARAM-VALUE %d34
// PARAM-NAME      = SD-NAME
// SD-NAME         = 1*32PRINTUSASCII except '=', SP, ']', %d34 (")
func readSdParamName(r io.RuneScanner) (string, error) {
	rv := &bytes.Buffer{}
	for {
		ch, _, err := r.ReadRune()
		if err != nil {
			return "", err
		}
		if ch == '=' {
			r.UnreadRune()
			return string(rv.Bytes()), nil
		}
		rv.WriteRune(ch)
	}
}

// readSdParamValue reads an PARAM-VALUE as defined by RFC-5424
// SD-PARAM        = PARAM-NAME "=" %d34 PARAM-VALUE %d34
// PARAM-VALUE     = UTF-8-STRING ; characters '"', '\' and ']' MUST be escaped.
func readSdParamValue(r io.RuneScanner) (string, error) {
	ch, _, err := r.ReadRune()
	if err != nil {
		return "", err
	}
	if ch != '"' {
		return "", badFormat("StructuredData[].Parameters[]") // hard to reach
	}

	rv := &bytes.Buffer{}
	for {
		ch, _, err := r.ReadRune()
		if err != nil {
			return "", err
		}
		if ch == '\\' {
			ch, _, err := r.ReadRune()
			if err != nil {
				return "", err
			}
			rv.WriteRune(ch)
			continue
		}
		if ch == '"' {
			return string(rv.Bytes()), nil
		}
		rv.WriteRune(ch)
	}
}

// readSpace reads a single space
func readSpace(r io.RuneScanner) error {
	ch, _, err := r.ReadRune()
	if err != nil {
		return err
	}
	if ch != ' ' {
		return badFormat("expected space")
	}
	return nil
}

// readWord reads r until it encounters a space (0x20)
func readWord(r io.RuneScanner) (string, error) {
	rv := &bytes.Buffer{}
	for {
		ch, _, err := r.ReadRune()
		if err != nil {
			return "", err
		} else if ch != ' ' {
			rv.WriteRune(ch)
		} else {
			r.UnreadRune()
			rvString := string(rv.Bytes())
			if rvString == "-" {
				rvString = ""
			}
			return rvString, nil
		}
	}
}

func copyFrom(in []byte) []byte {
	out := make([]byte, len(in))
	copy(out, in)
	return out
}

const severityMask = 0x07
const facilityMask = 0xf8

type Priority int

const (
	Emergency Priority = iota
	Alert
	Crit
	Error
	Warning
	Notice
	Info
	Debug
)

const (
	Kern Priority = iota << 3
	User
	Mail
	Daemon
	Auth
	Syslog
	Lpr
	News
	Uucp
	Cron
	Authpriv
	Ftp
	Local0
	Local1
	Local2
	Local3
	Local4
	Local5
	Local6
	Local7
)
`
