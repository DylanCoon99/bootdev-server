package main

import (
	"fmt"
	"net/http"
	"encoding/json"
	"strings"
	"time"
	"github.com/DylanCoon99/bootdev-server/internal/database"
	"github.com/DylanCoon99/bootdev-server/internal/auth"
	"github.com/google/uuid"

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
		Password string `json:password`
		Email    string `json:"email"`
	}

	var req request

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)

	if err != nil {
		// failed to decode request
		w.WriteHeader(500)
		return
	}

	hashed_password, err := auth.HashPassword(req.Password)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	email := req.Email


	params := database.CreateUserParams {
		Email:          email,
		HashedPassword: hashed_password,
	}


	user, err := cfg.DBQueries.CreateUser(r.Context(), params)
	if err != nil {
		// failed to create user
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	// encode the user to a json response

	b, err := json.Marshal(user)

	if err != nil {
		// failed to encode user to json
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(201)
	w.Write(b)
	return

}


func validateChirp(body string) (statusCode int, cleaned_body string) {

	// validate the chirp
	char_count := 0

	for _ = range body {
		char_count += 1
	}

	if char_count > 140 {
		// chirp is too long

		// bad request
		return 400, ""
	}

	// otherwise request is validated, build a response and encode it (marshal)

	cleaned_body = strings.ReplaceAll(body, "kerfuffle", "****")
	cleaned_body = strings.ReplaceAll(cleaned_body, "sharbert", "****")
	cleaned_body = strings.ReplaceAll(cleaned_body, "fornax", "****")
	cleaned_body = strings.ReplaceAll(cleaned_body, "Fornax", "****")


	return 200, cleaned_body

}



func (cfg *apiConfig) createChirpHandler(w http.ResponseWriter, r *http.Request) {

	// get the email from the request

	type request struct {
		Body   string    `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}

	var req request


	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("Couldn't find JWT"))
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		w.WriteHeader(401)
		w.Write([]byte("Couldn't validate JWT"))
		return
	}

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&req)

	if err != nil {
		// failed to decode request
		w.WriteHeader(500)
		w.Write([]byte("Failed to decode the request"))
		return
	}

	body := req.Body
	//user_id := req.UserId


	// validate the body
	status, cleaned_body := validateChirp(body)
	if status != 200 {
		// chirp was invalid

		w.WriteHeader(400)
		w.Write([]byte("Chirp is invalid"))
		return
	}


	params := database.CreateChirpParams {
		Body: cleaned_body,
		UserID: userID,
	}


	chirp, err := cfg.DBQueries.CreateChirp(r.Context(), params)
	if err != nil {
		// failed to create chirp
		w.WriteHeader(500)
		//fmt.Errorf("Failed to create the chirp", err)
		w.Write([]byte(err.Error()))
		return
	}

	// encode the user to a json response

	b, err := json.Marshal(chirp)

	if err != nil {
		// failed to encode user to json
		w.WriteHeader(500)
		w.Write([]byte("Failed to encode user to json"))
		return
	}

	w.WriteHeader(201)
	w.Write(b)
	return

}

func reverse(list []database.Chirp) []database.Chirp {
    for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
        list[i], list[j] = list[j], list[i]
    }
    return list
}



func (cfg *apiConfig) getChirpsHandler(w http.ResponseWriter, r *http.Request) {

	s := r.URL.Query().Get("author_id")
	sortParam := r.URL.Query().Get("sort")

	userID, err := uuid.Parse(s)
	if s != "" {
		if err != nil {
			w.WriteHeader(404)
			w.Write([]byte("Error parsing uuid"))
			return
		}

	}

	// declare a list of chirp structs

	var allChirpList []database.Chirp


	// call the query function to get all chirps from the database
	allChirpList, err = cfg.DBQueries.ListChirps(r.Context())

	if err != nil {
		w.WriteHeader(500) // database failed to get chirps
		return
	}


	chirpList := []database.Chirp{}

	for _,dbChirp := range allChirpList {
		if userID != uuid.Nil && dbChirp.UserID != userID {
			continue
		}

		chirpList = append(chirpList, database.Chirp{
			ChirpID:   dbChirp.ChirpID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			UserID:    dbChirp.UserID,
			Body:      dbChirp.Body,
		})
	}


	if sortParam == "" || sortParam == "asc" {
		b, _ := json.Marshal(chirpList)

		w.WriteHeader(200)
		w.Write(b)
		return
	} 

	chirpList = reverse(chirpList)
	b, _ := json.Marshal(chirpList)

	w.WriteHeader(200)
	w.Write(b)
	return

}



func (cfg *apiConfig) getChirpHandler(w http.ResponseWriter, r *http.Request) {

	chirp_id, err := uuid.Parse(r.PathValue("chirpID"))

	// query the database for a specific chirp

	chirp, err := cfg.DBQueries.GetChirp(r.Context(), chirp_id)

	if err != nil {
		w.WriteHeader(404) // failed to query database
		w.Write([]byte(err.Error()))
		return
	}

	b, err := json.Marshal(chirp)

	w.WriteHeader(200)
	w.Write(b)

}


func (cfg apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {

	chirp_id, err := uuid.Parse(r.PathValue("chirpID"))

	// authenticate the user

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(401)
		w.Write([]byte("Couldn't find JWT"))
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		w.WriteHeader(403)
		w.Write([]byte("Couldn't validate JWT"))
		return
	}

	// check if the chirp belongs to this user.
	chirp, err := cfg.DBQueries.GetChirp(r.Context(), chirp_id)
	if chirp.UserID != userID {
		w.WriteHeader(403)
		w.Write([]byte("Authentication failed"))
		return
	}


	// then delete the chirp
	chirp, err = cfg.DBQueries.DeleteChirp(r.Context(), chirp_id)

	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte(err.Error()))
		return
	}


	b, err := json.Marshal(chirp)

	w.WriteHeader(204)
	w.Write(b)
}



func (cfg apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {


	type request struct {
		Password          string `json:"password"`
		Email             string `json:"email"`
		ExpiresInSeconds  int    `json:"expires_in_seconds"`
	}


	var req request

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)

	email := req.Email

	// query the database to see if the users email matches the password provided
	user, err := cfg.DBQueries.GetUser(r.Context(), email)

	if err != nil {
		w.WriteHeader(401) // failed to query database
		w.Write([]byte("Incorrect email or password"))
		return
	}


	// check the hashed password from the user
	hashed_password := user.HashedPassword

	err = auth.CheckPassword(req.Password, hashed_password)

	if err != nil {
		w.WriteHeader(401) // failed to query database
		w.Write([]byte("Incorrect email or password"))
		return
	}


	type response struct {
		User struct {
			Id          string       `json:"id"`
			CreatedAt   string       `json:"created_at"`
			UpdatedAt   string       `json:"updated_at"`
		}
		Email         string `json:"email"`
		IsChirpyRed   bool   `json:"is_chirpy_red"`
		Token         string `json:"token"`
		RefreshToken  string `json:"refresh_token"`
	}



	expirationTime := time.Hour
	if req.ExpiresInSeconds > 0 && req.ExpiresInSeconds < 3600 {
		expirationTime = time.Duration(req.ExpiresInSeconds) * time.Second
	}

	accessToken, err := auth.MakeJWT(
		user.ID,
		cfg.jwtSecret,
		expirationTime,
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Couldn't create access JWT"))
		return
	}

	refresh_token, err := auth.MakeRefreshToken()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Couldn't create refresh token"))
		return
	}
	_, err = cfg.DBQueries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams {
		UserID:    user.ID,
		Token:     refresh_token,
		ExpiresAt: time.Now().UTC().Add(time.Hour * 24 * 60),
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Couldn't create refresh token"))
	}


	var res response

	res.User.Id          = (user.ID).String()
	res.User.CreatedAt   = (user.CreatedAt).String()
	res.User.UpdatedAt   = (user.UpdatedAt).String()
	res.Email       = email
	res.IsChirpyRed = user.IsChirpyRed
	res.Token            = accessToken
	res.RefreshToken     = refresh_token

	b, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(500) // failed to query database
		w.Write([]byte("Failed to encode json"))
		return
	}

	w.WriteHeader(200)
	w.Write(b)

}


// "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJjaGlycHkiLCJzdWIiOiIyMTVhNDZmZS1jMDcwLTRiOGYtOWEwMy03OTZkMTJhMzkwOTAiLCJleHAiOjE3MzE1NjA1NzIsImlhdCI6MTczMTU1Njk3Mn0.Kb_pqnuJotRGlIiyIQsWDCUkTAvOjRBCIe3DambMPUc"

// "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJjaGlycHkiLCJzdWIiOiJiYjQ2OTA2Ni02YTc0LTRiZGItYTQ1Ni1lYjk1MmU5NzU3ODciLCJleHAiOjE3MzE1NjA1OTQsImlhdCI6MTczMTU1Njk5NH0.eZ8KvreDMrIIfiQhIWJXAc71Xa5lSXqr8PJAwU7aWLo"



func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}

	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Couldn't find token"))
		return
	}

	user, err := cfg.DBQueries.GetUserFromRefreshToken(r.Context(), refreshToken)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Couldn't get user for refresh token"))
		return
	}

	accessToken, err := auth.MakeJWT(
		user.ID,
		cfg.jwtSecret,
		time.Hour,
	)
	if err != nil {
		//respondWithError(w, http.StatusUnauthorized, "Couldn't validate token", err)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Couldn't validate token"))		
		return
	}


	var res response

	res.Token = accessToken

	// marshal the response


	b, err := json.Marshal(res)

	w.WriteHeader(http.StatusOK)
	w.Write(b)
}







func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Couldn't find token"))
		return
	}

	_, err = cfg.DBQueries.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Couldn't revoke session"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}




func (cfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {

	// users can only change their own email and password. user must be validated

	type request struct {
		Email string
		Password string
	}

	token, err := auth.GetBearerToken(r.Header)

	if err != nil {
		w.WriteHeader(401)
		w.Write([]byte("Couldn't find JWT"))
		return
	}

	var req request

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&req)

	if err != nil {
		// failed to decode request
		w.WriteHeader(500)
		w.Write([]byte("Failed to decode the request"))
		return
	}



	

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		w.WriteHeader(401)
		w.Write([]byte("Couldn't validate JWT"))
		return
	}


	hashed_password, err := auth.HashPassword(req.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to hash password"))
	}


	// build update user params
	params := database.UpdateUserParams {
		Email: req.Email,
		HashedPassword: hashed_password,
		ID: userID,
	}


	user, err := cfg.DBQueries.UpdateUser(r.Context(), params)
	if err != nil {
		w.WriteHeader(401)
		w.Write([]byte("Failed to update user"))
	}


	type response struct {
		ID             uuid.UUID    `json:"id"`
		CreatedAt      time.Time    `json:"created_at"`
		UpdatedAt      time.Time    `json:"updated_at"`
		Email          string       `json:"email"`
		IsChirpyRed    bool         `json:"is_chirpy_red"`
	}

	var res response

	res.ID        = user.ID
	res.Email     = user.Email
	res.CreatedAt = user.CreatedAt
	res.UpdatedAt = user.UpdatedAt


	b, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to encode json response"))
	}


	w.WriteHeader(http.StatusOK)
	w.Write(b)

	return
}



func (cfg *apiConfig) handlerWebhook(w http.ResponseWriter, r *http.Request) {

	apiKey, _ := auth.GetAPIKey(r.Header)

	if apiKey != cfg.polkaKey {
		w.WriteHeader(401)
		//w.Write()
		return
	}

	type request struct {
		Event string `json:"event"`
		Data struct {
			UserID uuid.UUID `json:"user_id"`
		}
	}

	var req request

	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&req)


	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to decode request"))
		return
	}

	if req.Event != "user.upgraded" {
		w.WriteHeader(204)
		w.Write([]byte("Not user upgraded"))
		return
	}

	// if user.upgraded --> upgrade the user in the database
	user, err := cfg.DBQueries.UpgradeChirpyRed(r.Context(), req.Data.UserID)
	if user.ID != req.Data.UserID {
		w.WriteHeader(404)
		w.Write([]byte("User not found"))
		return
	}
	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte((req.Data.UserID).String()))
		return
	}

	w.WriteHeader(204)
	w.Write([]byte("Upgrade successful"))
}