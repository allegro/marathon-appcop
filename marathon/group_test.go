package marathon

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseGroupsWhenMalformedJSONSIPassedShouldReturnErrorAndNilGroups(t *testing.T) {
	t.Parallel()
	// given
	var jsonBlob = []byte(`{
      "groups": [
  }
	`)

	// when
	groups, err := ParseGroups(jsonBlob)

	// then
	require.Error(t, err)
	assert.Nil(t, groups)

}

func TestParseGroupsWhenProperJSONIsProvidedWithGroupAndEmptyAppsSholdReturnGroupAndNoError(t *testing.T) {
	t.Parallel()
	// given
	var jsonBlob = []byte(`{
      "groups": [
        {
          "apps": [],
          "dependencies": [],
          "groups": [],
          "id": "/com.example.tech.maas",
          "version": "2017-01-24T15:37:58.780Z"
        }
      ]
  }
	`)

	expectedGroups := []*Group{
		{Apps: []*App{},
			Groups:  []*Group{},
			ID:      "/com.example.tech.maas",
			Version: "2017-01-24T15:37:58.780Z"},
	}
	// when
	groups, err := ParseGroups(jsonBlob)

	// then
	require.NoError(t, err)
	assert.Equal(t, expectedGroups, groups)

}
