package auth

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CodetokenExchanger", func() {
	Describe("newExchangeCodeRequest", func() {
		It("creates the request", func() {
			tokenRetriever := TokenRetriever{oidcWellKnownEndpoints: OIDCWellKnownEndpoints{TokenEndpoint: "https://issuer/oauth/token"}}
			exchangeRequest := AuthorizationCodeExchangeRequest{
				ClientID:     "clientID",
				CodeVerifier: "Verifier",
				Code:         "code",
				RedirectURI:  "https://redirect",
			}

			result, err := tokenRetriever.newExchangeCodeRequest(exchangeRequest)

			var tokenRequest AuthorizationTokenRequest
			json.NewDecoder(result.Body).Decode(&tokenRequest)

			Expect(err).To(BeNil())
			Expect(tokenRequest).To(Equal(AuthorizationTokenRequest{
				GrantType:    "authorization_code",
				ClientID:     "clientID",
				CodeVerifier: "Verifier",
				Code:         "code",
				RedirectURI:  "https://redirect",
			}))
			Expect(result.URL.String()).To(Equal("https://issuer/oauth/token"))
		})

		It("returns an error when NewRequest returns an error", func() {
			tokenRetriever := TokenRetriever{oidcWellKnownEndpoints: OIDCWellKnownEndpoints{TokenEndpoint: "://issuer/oauth/token"}}

			result, err := tokenRetriever.newExchangeCodeRequest(AuthorizationCodeExchangeRequest{})

			Expect(result).To(BeNil())
			Expect(err.Error()).To(Equal("parse ://issuer/oauth/token: missing protocol scheme"))
		})
	})

	Describe("handleExhcangeCodeResponse", func() {
		It("handles the response", func() {
			tokenRetriever := TokenRetriever{}
			response := buildResponse(200, AuthorizationTokenResponse{
				ExpiresIn:    1,
				AccessToken:  "myAccessToken",
				RefreshToken: "myRefreshToken",
			})

			result, err := tokenRetriever.handleAuthTokensResponse(response)

			Expect(err).To(BeNil())
			Expect(result).To(Equal(&TokenResult{
				ExpiresIn:    1,
				AccessToken:  "myAccessToken",
				RefreshToken: "myRefreshToken",
			}))
		})

		It("returns error when status code is not successful", func() {
			tokenRetriever := TokenRetriever{}
			response := buildResponse(500, nil)

			result, err := tokenRetriever.handleAuthTokensResponse(response)

			Expect(result).To(BeNil())
			Expect(err.Error()).To(Equal("A non-success status code was receveived: 500"))
		})

		It("returns error when deserialization fails", func() {
			tokenRetriever := TokenRetriever{}
			response := buildResponse(200, "")

			result, err := tokenRetriever.handleAuthTokensResponse(response)
			Expect(result).To(BeNil())
			Expect(err.Error()).To(Equal("json: cannot unmarshal string into Go value of type auth.AuthorizationTokenResponse"))
		})
	})

	Describe("newRefreshTokenRequest", func() {
		It("creates the request", func() {
			tokenRetriever := TokenRetriever{oidcWellKnownEndpoints: OIDCWellKnownEndpoints{TokenEndpoint: "https://issuer/oauth/token"}}
			exchangeRequest := RefreshTokenExchangeRequest{
				ClientID:     "clientID",
				RefreshToken: "refreshToken",
			}

			result, err := tokenRetriever.newRefreshTokenRequest(exchangeRequest)

			var tokenRequest RefreshTokenRequest
			json.NewDecoder(result.Body).Decode(&tokenRequest)

			Expect(err).To(BeNil())
			Expect(tokenRequest).To(Equal(RefreshTokenRequest{
				GrantType:    "refresh_token",
				ClientID:     "clientID",
				RefreshToken: "refreshToken",
			}))
			Expect(result.URL.String()).To(Equal("https://issuer/oauth/token"))
		})

		It("returns an error when NewRequest returns an error", func() {
			tokenRetriever := TokenRetriever{oidcWellKnownEndpoints: OIDCWellKnownEndpoints{TokenEndpoint: "://issuer/oauth/token"}}

			result, err := tokenRetriever.newRefreshTokenRequest(RefreshTokenExchangeRequest{})

			Expect(result).To(BeNil())
			Expect(err.Error()).To(Equal("parse ://issuer/oauth/token: missing protocol scheme"))
		})
	})
})

func buildResponse(statusCode int, body interface{}) *http.Response {
	b, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}

	resp := &http.Response{
		StatusCode: statusCode,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte(b))),
	}

	return resp
}
