package spotify

import (
	"chatserver/data"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

const (
	redirectURI  = "http://localhost:8100/callback"
	authorizeURL = "https://accounts.spotify.com/authorize/?"
	tokenURL     = "https://accounts.spotify.com/api/token"
)

func Login(w http.ResponseWriter, r *http.Request) {
	parsedURL, err := url.Parse(authorizeURL)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Println("Error parsing URL:", err)
		return
	}

	queryParams := parsedURL.Query()
	queryParams.Add("client_id", os.Getenv("SPOTIFY_CLIENT_ID"))
	queryParams.Add("redirect_uri", redirectURI)
	queryParams.Add("response_type", "code")
	queryParams.Add("scope", `streaming user-library-modify user-library-read user-modify-playback-state user-read-playback-state user-read-email user-read-private playlist-read-private`)

	state, err := gonanoid.New()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Println("Error generating state:", err)
		return
	}
	queryParams.Add("state", state)

	parsedURL.RawQuery = queryParams.Encode()

	setJSONResponseHeaders(w)
	response := data.Message{Data: parsedURL.String()}
	json.NewEncoder(w).Encode(response)
}

func Callback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if state == "" {
		writeErrorResponse(w, http.StatusBadRequest, "State Mismatch")
		return
	}

	formData := url.Values{
		"code":         {code},
		"redirect_uri": {redirectURI},
		"grant_type":   {"authorization_code"},
	}

	payload := strings.NewReader(formData.Encode())

	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	headers := map[string]string{
		"Content-Type":  "application/x-www-form-urlencoded",
		"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(clientID+":"+clientSecret)),
	}

	req, err := http.NewRequest("POST", tokenURL, payload)
	if err != nil {
		writeInternalServerError(w, "Error creating request", err)
		return
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		writeInternalServerError(w, "Error sending request", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		writeInternalServerError(w, fmt.Sprintf("Unexpected response status code: %d", resp.StatusCode), nil)
		return
	}

	var responseBody map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		writeInternalServerError(w, "Error decoding response", err)
		return
	}

	setJSONResponseHeaders(w)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseBody)
}

func Refresh(w http.ResponseWriter, r *http.Request) {
	refreshToken := r.URL.Query().Get("refresh_token")

	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")

	formData := url.Values{
		"client_id":     {clientID},
		"refresh_token": {refreshToken},
		"grant_type":    {"refresh_token"},
	}

	payload := strings.NewReader(formData.Encode())

	headers := map[string]string{
		"Content-Type":  "application/x-www-form-urlencoded",
		"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(clientID+":"+clientSecret)),
	}

	req, err := http.NewRequest("POST", tokenURL, payload)
	if err != nil {
		writeInternalServerError(w, "Error creating request", err)
		return
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		writeInternalServerError(w, "Error sending request", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		writeInternalServerError(w, fmt.Sprintf("Unexpected response status code: %d", resp.StatusCode), nil)
		return
	}

	var responseBody map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		writeInternalServerError(w, "Error decoding response", err)
		return
	}

	setJSONResponseHeaders(w)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseBody)

}

func writeInternalServerError(w http.ResponseWriter, message string, err error) {
	if err != nil {
		log.Println(message+":", err)
	} else {
		log.Println(message)
	}
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}

func writeErrorResponse(w http.ResponseWriter, statusCode int, errorMessage string) {
	w.WriteHeader(statusCode)
	setJSONResponseHeaders(w)
	response := data.ErrorResponse{Error: errorMessage}
	json.NewEncoder(w).Encode(response)
}

func setJSONResponseHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
}
