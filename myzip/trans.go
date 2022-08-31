package myzip

import (
	"bytes"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
	"strings"
	//  "code.google.com/p/go.text/transform" // deprecated
	"golang.org/x/text/encoding/japanese"
	//  "code.google.com/p/go.text/encoding/japanese" // deprecated
)

const (
	//simple_chinese
	GB18030  = "GB18030"
	GBK      = "GBK"
	HZGB2312 = "HZGB2312"

	//tz
	Big5 = "Big5"

	//Japanese
	EUCJP     = "EUCJP"
	ISO2022JP = "ISO2022JP"
	ShiftJIS  = "ShiftJIS"

	//korean
	EUCKR = "EUCKR"

	//basic
	UTF8 = "UTF8"
)

var TransFormMap = map[string]encoding.Encoding{
	GB18030:   simplifiedchinese.GB18030,
	GBK:       simplifiedchinese.GBK,
	HZGB2312:  simplifiedchinese.HZGB2312,
	Big5:      traditionalchinese.Big5,
	EUCJP:     japanese.EUCJP,
	ISO2022JP: japanese.ISO2022JP,
	ShiftJIS:  japanese.ShiftJIS,
	EUCKR:     korean.EUCKR,
	UTF8:      unicode.UTF8,
}

func FindTransForm(transformName string) encoding.Encoding {
	if transformName == "" {
		return TransFormMap[UTF8]
	}
	value, isExist := TransFormMap[transformName]
	if !isExist {
		return TransFormMap[UTF8]
	}
	return value
}

func transformEncoding(rawReader io.Reader, trans transform.Transformer) (string, error) {
	ret, err := ioutil.ReadAll(transform.NewReader(rawReader, trans))
	if err == nil {
		return string(ret), nil
	} else {
		return "", err
	}
}

//FromShiftJIS  Convert a string encoding from ShiftJIS to UTF-8
//https://gist.github.com/hyamamoto/db03c03fd624881d4b84
func FromXXToUTF8(input string, charsetName string) (string, error) {
	form := FindTransForm(charsetName)
	return transformEncoding(strings.NewReader(input), form.NewDecoder())
}

// Convert a string encoding from UTF-8 to ShiftJIS
func FromUTF8ToXX(input string, charsetName string) (string, error) {
	form := FindTransForm(charsetName)
	return transformEncoding(strings.NewReader(input), form.NewEncoder())
}

// Convert an array of bytes (a valid ShiftJIS string) to a UTF-8 string
func BytesFromXXToUTF8(input []byte, charsetName string) (string, error) {
	form := FindTransForm(charsetName)
	return transformEncoding(bytes.NewReader(input), form.NewDecoder())
}

// Convert an array of bytes (a valid UTF-8 string) to a ShiftJIS string
func BytesFromUTF8ToXX(input []byte, charsetName string) (string, error) {
	form := FindTransForm(charsetName)
	return transformEncoding(bytes.NewReader(input), form.NewEncoder())
}

