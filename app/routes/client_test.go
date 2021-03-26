package routes_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"oauth2-server/app/domain"
	"oauth2-server/app/routes"
	"oauth2-server/app/test"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateClientSuccess(t *testing.T) {
	var server = test.TestServerFor(routes.ClienRouter)

	clientData := map[string]interface{}{
		"client_id":             "new-test-client",
		"secret":                "secret",
		"allowed_redirect_urls": []string{"http://test.com/callback"},
		"allowed_grant_types":   []string{"code"},
		"allowed_scopes":        []string{"profile"},
	}

	jsonValue, _ := json.Marshal(clientData)
	req, _ := http.NewRequest("POST", server.URL+"/oauth2/client", bytes.NewBuffer(jsonValue))

	resp, _ := httpClient().Do(req)
	var respBody domain.OAuthError

	json.NewDecoder(resp.Body).Decode(&respBody)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}
