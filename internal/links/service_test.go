package links

import (
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/stretchr/testify/assert"
)

func TestValidateURL(t *testing.T) {
	cases := []struct {
		name, rawURL string
		wantErr      bool
		targetErr    error
	}{
		{"validURL", "https://example.com", false, nil},
		{"validURLWithPath", "https://example.com/path", false, nil},
		{"invalidEmptyURL", "", true, ErrEmptyURL},
		{"invalidSchemeURL", "ftp://example.com", true, ErrUnsupportScheme},
		{"invalidEmptyHostnameURL", "https://", true, ErrEmptyHostname},
		{"invalidEncodingURL", "https://example.com/%zz", true, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			err := validateURL(c.rawURL)
			if c.wantErr {
				assert.Error(t, err)
				if c.targetErr != nil {
					assert.ErrorIs(t, err, c.targetErr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
