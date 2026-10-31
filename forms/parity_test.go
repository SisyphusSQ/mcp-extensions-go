package forms_test

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/SisyphusSQ/mcp-extensions-go/forms"
)

// These expectations are produced by the official Python extension, not by Go.
func TestPythonProtocolParity(t *testing.T) {
	data, err := os.ReadFile("testdata/python-parity.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Commit       string `json:"upstream_commit"`
		SourceSHA256 string `json:"upstream_source_sha256"`
		Cases        []struct {
			Name   string
			Schema json.RawMessage
			Valid  bool
			Values []struct {
				Value json.RawMessage
				Valid bool
			}
			Uploads []struct {
				Content  map[string]any
				Uploaded []string
				Valid    bool
				Result   json.RawMessage
			}
		}
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Commit != "36f8b38df179144b44d157133bbc5f7c6e353ace" || fixture.SourceSHA256 != "01e79d6a3bf1aaa14d7b6cd368dabb6fb53c71f56f330ee0875608f3d72a7c66" || len(fixture.Cases) == 0 {
		t.Fatal("missing reference provenance")
	}
	for _, row := range fixture.Cases {
		t.Run(row.Name, func(t *testing.T) {
			form, err := forms.Parse(row.Schema)
			if (err == nil) != row.Valid {
				t.Fatalf("declaration valid=%v, Python=%v: %v", err == nil, row.Valid, err)
			}
			if !row.Valid {
				return
			}
			for _, answer := range row.Values {
				err := form.ValidateField("value", answer.Value, nil)
				if (err == nil) != answer.Valid {
					t.Errorf("value %s valid=%v, Python=%v: %v", answer.Value, err == nil, answer.Valid, err)
				}
			}
			for _, upload := range row.Uploads {
				result, err := form.CompleteSubmission("value", upload.Content, upload.Uploaded, nil)
				if (err == nil) != upload.Valid {
					t.Errorf("upload %v valid=%v, Python=%v: %v", upload.Uploaded, err == nil, upload.Valid, err)
					continue
				}
				if upload.Valid {
					encoded, _ := json.Marshal(result)
					var got, want any
					_ = json.Unmarshal(encoded, &got)
					_ = json.Unmarshal(upload.Result, &want)
					if !reflect.DeepEqual(got, want) {
						t.Errorf("upload result=%s, Python=%s", encoded, upload.Result)
					}
				}
			}
		})
	}
}
