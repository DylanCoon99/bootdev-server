package main

import (
	"fmt"
	"net/http"
	"encoding/json"
	"strings"
)


func healthHandler(w http.ResponseWriter, r *http.Request) {
	// This is the handler for the /healthz endpoint
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}


func (cfg *apiConfig) hitsHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)


	hits := int(cfg.fileserverHits.Load())

	html := fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, hits)


	fmt.Fprint(w, html)

	//w.Write([]byte(fmt.Sprintf("Hits: %d", hits)))
}


func (cfg *apiConfig) resetMetricsHandler(w http.ResponseWriter, r *http.Request) {
	
	platform := cfg.Platform

	if platform != "dev" {
		// forbidden 
		w.WriteHeader(http.StatusForbidden)
		return
	}


	cfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0"))
}




func (cfg *apiConfig) createUserHandler(w http.ResponseWriter, r *http.Request) {

	// get the email from the request

	type request struct {
		Email string `json:"email"`
	}

	var req request

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)

	if err != nil {
		// failed to decode request
		w.WriteHeader(500)
		return
	}

	email := req.Email


	user, err := cfg.DBQueries.CreateUser(r.Context(), email)
	if err != nil {
		// failed to create user
		w.WriteHeader(500)
		return
	}

	// encode the user to a json response

	b, err := json.Marshal(user)

	if err != nil {
		// failed to encode user to json
		return
	}

	w.WriteHeader(201)
	w.Write(b)
	return

}


func validateChirp(w http.ResponseWriter, r *http.Request) {

	// decode the json from the header

	type request struct {
		Body string `json:"body"`
	}


	var req request


	decoder := json.NewDecoder(r.Body) // create a new decoder for the request body
	err := decoder.Decode(&req)


	type errorResponse struct {
		Error string `json:"error"`
	}


	if err != nil {
		fmt.Errorf("Error decoding request: %s", err)

		eRes := errorResponse {
			Error: "Error decoding request",
		}

		b, _ := json.Marshal(eRes)

		// write bytes to the header

		w.WriteHeader(500)
		w.Write([]byte(b))
		return
	}


	// validate the chirp
	body := req.Body
	char_count := 0

	for _ = range body {
		char_count += 1
	}

	if char_count > 140 {

		eRes := errorResponse {
			Error: "Chirp is too long",
		}

		b, _ := json.Marshal(eRes)

		// write bytes to the header
		w.WriteHeader(400)
		w.Write([]byte(b))
		 // bad request
		return
	}

	// otherwise request is validated, build a response and encode it (marshal)
	type response struct {
		CleanedBody string `json:"cleaned_body"`
	}

	cleaned_body := strings.ReplaceAll(body, "kerfuffle", "****")
	cleaned_body = strings.ReplaceAll(cleaned_body, "sharbert", "****")
	cleaned_body = strings.ReplaceAll(cleaned_body, "fornax", "****")
	cleaned_body = strings.ReplaceAll(cleaned_body, "Fornax", "****")


	res := response {
		CleanedBody: cleaned_body,
	}


	b, _ := json.Marshal(res)

	w.WriteHeader(200) // good request
	w.Write([]byte(b))


}



func (cfg *apiConfig) createChirpHandler(w http.ResponseWriter, r *http.Request) {

	// get the email from the request

	type request struct {
		Body string   `json:"body"`
		UserId string `json:"user_id"`
	}

	var req request

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)

	if err != nil {
		// failed to decode request
		w.WriteHeader(500)
		return
	}

	body := req.Bodys
	user_id := req.UserId


	// validate the body
	


	chirp, err := cfg.DBQueries.CreateChirp(r.Context(), body, user_id)
	if err != nil {
		// failed to create chirp
		w.WriteHeader(500)
		return
	}

	// encode the user to a json response

	b, err := json.Marshal(chirp)

	if err != nil {
		// failed to encode user to json
		return
	}

	w.WriteHeader(201)
	w.Write(b)
	return

}