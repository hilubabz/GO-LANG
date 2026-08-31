package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"

	"github.com/hilubabz/GO-LANG/internal/auth"
	"github.com/hilubabz/GO-LANG/internal/database"
	"github.com/hilubabz/GO-LANG/middleware"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platform string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	hits := cfg.fileserverHits.Load()
	w.Write([]byte(fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, hits)))
}

func (cfg *apiConfig) addUser(w http.ResponseWriter, r *http.Request){
	type userRequest struct{
		Password string `json:"password"`
		Email string `json:"email"`
	}
	user := userRequest{}
	err := json.NewDecoder(r.Body).Decode(&user)
	if err!=nil{
		respondWithError(w, http.StatusBadRequest, "Unable to parse body")
		return
	}
	hashedPassword, passwordError := auth.HashPassword(user.Password)
	if passwordError != nil{
		respondWithError(w, http.StatusBadRequest, passwordError.Error())
	}
	userRes, insError := cfg.dbQueries.CreateUser(r.Context(), database.CreateUserParams{
		Email: user.Email,
		HashedPassword: hashedPassword,
	})
	if insError!=nil{
		respondWithError(w, http.StatusBadRequest, "Failed to add user")
		return
	}
	type userResponse struct{
		ID string `json:"id"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		Email string `json:"email"`
	}
	response := userResponse{
		ID: userRes.ID.String(),
		CreatedAt: userRes.CreatedAt.GoString(),
		UpdatedAt: userRes.UpdatedAt.GoString(),
		Email: userRes.Email,
	}
	respondWithJSON(w, http.StatusCreated, response)
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	if cfg.platform!="dev"{
		respondWithError(w, http.StatusForbidden, "This can only be accessed in dev mode")
		return 
	}
	err:=cfg.dbQueries.DeleteUser(r.Context())
	if err!=nil{
		respondWithError(w, http.StatusBadRequest, "Failed to delete users")
		return
	}
	cfg.fileserverHits.Store(0)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("User deleted successfully"))
}

func (cfg *apiConfig) validateChirp(w http.ResponseWriter, r *http.Request) {
	token, tokenErr := auth.GetBearerToken(r.Header)
	if tokenErr!=nil{
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	key, keyErr := os.LookupEnv("JWT_SECRET")
	if !keyErr{
		respondWithError(w, http.StatusBadRequest, "Key not found")
		return
	}
	userId, validErr := auth.ValidateJWT(token, key)
	if validErr!=nil{
		respondWithError(w, http.StatusUnauthorized, validErr.Error())
		return
	}
	type validateChirpBody struct {
		Body string `json:"body"`
	}

	validator := validateChirpBody{}

	err := json.NewDecoder(r.Body).Decode(&validator)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Something went wrong")
		return
	}

	if len(validator.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}
	validator.Body=strings.ReplaceAll(validator.Body,"kerfuffle","****")
	validator.Body=strings.ReplaceAll(validator.Body,"Kerfuffle","****")
	validator.Body=strings.ReplaceAll(validator.Body,"fornax","****")
	validator.Body=strings.ReplaceAll(validator.Body,"Fornax","****")
	validator.Body=strings.ReplaceAll(validator.Body,"sharbert","****")
	validator.Body=strings.ReplaceAll(validator.Body,"Sharbert","****")

	chirpData, addError := cfg.dbQueries.AddChirp(r.Context(), database.AddChirpParams{
		Body: validator.Body,
		UserID: userId,
	})

	if addError != nil{
		respondWithError(w, http.StatusBadRequest, "Failed to add chirp")
	}

	type validateChirpResponse struct {
		ID uuid.UUID `json:"id"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		Body string `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}

	response := validateChirpResponse{
		ID: chirpData.ID,
		CreatedAt: chirpData.CreatedAt.GoString(),
		UpdatedAt: chirpData.UpdatedAt.GoString(),
		Body: chirpData.Body,
		UserId: chirpData.UserID,
	}

	respondWithJSON(w, http.StatusCreated, response)
}

func (cfg *apiConfig) getChirps(w http.ResponseWriter, r *http.Request){
	chirpsData, err := cfg.dbQueries.GetChirps(r.Context())
	if err!=nil{
		respondWithError(w, http.StatusBadRequest, "Failed to fetch chirps")
	}
	type chirpsResponseType struct{
		ID uuid.UUID `json:"id"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		Body string `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}
	chirpsRes := []chirpsResponseType{}

	for _, chirp := range chirpsData {
		chirpsRes = append(chirpsRes, chirpsResponseType{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt.GoString(),
			UpdatedAt: chirp.UpdatedAt.GoString(),
			Body:      chirp.Body,
			UserId:    chirp.UserID,
		})
	}

	respondWithJSON(w, http.StatusOK, chirpsRes)
}

func (cfg *apiConfig) getChirpById(w http.ResponseWriter, r *http.Request){
	chirpId := r.PathValue("chirpId")
	chirpUuid, parseError := uuid.Parse(chirpId)
	if parseError != nil{
		respondWithError(w, http.StatusBadRequest, "Could not parsed chirp id")
	}
	chirpData, err := cfg.dbQueries.GetChirpById(r.Context(), chirpUuid)
	if err!=nil {
		respondWithError(w, http.StatusNotFound, "Failed to get by id")
		return
	}
	type chirpResponseType struct{
		ID uuid.UUID `json:"id"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		Body string `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}
	chirpResponse := chirpResponseType{
		ID: chirpData.ID,
		CreatedAt: chirpData.CreatedAt.GoString(),
		UpdatedAt: chirpData.UpdatedAt.GoString(),
		Body: chirpData.Body,
		UserId: chirpData.UserID,
	}
	respondWithJSON(w, http.StatusOK, chirpResponse)
}

func (cfg *apiConfig) login(w http.ResponseWriter, r *http.Request) {
	type loginRequest struct {
		Email   string `json:"email"`
		Password string `json:"password"`
	}

	loginData := loginRequest{}

	err := json.NewDecoder(r.Body).Decode(&loginData)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse json")
		return
	}

	userData, userError := cfg.dbQueries.GetUserByEmail(
		r.Context(),
		loginData.Email,
	)
	if userError != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	passwordMatch, passwordError := auth.CheckPasswordHash(
		loginData.Password,
		userData.HashedPassword,
	)
	if passwordError != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to compare password")
		return
	}

	if !passwordMatch {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	key, keyExists := os.LookupEnv("JWT_SECRET")
	if !keyExists {
		respondWithError(w, http.StatusInternalServerError, "JWT secret not configured")
		return
	}

	token, tokenErr := auth.MakeJWT(
		userData.ID,
		key,
	)
	if tokenErr != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create token")
		return
	}
	refreshToken := auth.MakeRefreshToken()

	refToken, refError := cfg.dbQueries.AddRefreshToken(r.Context(),database.AddRefreshTokenParams{
		Token: refreshToken,
		UserID: userData.ID,
		ExpiresAt: time.Now().Add(time.Duration(24*60)*time.Hour),
	})

	if refError!=nil{
		respondWithError(w, http.StatusBadRequest, refError.Error())
		return
	}

	type loginResponse struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt string    `json:"created_at"`
		UpdatedAt string    `json:"updated_at"`
		Email     string    `json:"email"`
		Token     string    `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	loginRes := loginResponse{
		ID:        userData.ID,
		CreatedAt: userData.CreatedAt.GoString(),
		UpdatedAt: userData.UpdatedAt.GoString(),
		Email:     userData.Email,
		Token:     token,
		RefreshToken: refToken.Token,
	}

	respondWithJSON(w, http.StatusOK, loginRes)
}

func (cfg *apiConfig) refresh(w http.ResponseWriter, r *http.Request){
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err!=nil{
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}
	validRefreshToken, validErr := cfg.dbQueries.GetRefreshToken(r.Context(),refreshToken)
	if validErr != nil{
		respondWithError(w, http.StatusUnauthorized, validErr.Error())
		return
	}
	key, _ := os.LookupEnv("JWT_SECRET")
	accessToken, tokenErr := auth.MakeJWT(validRefreshToken.UserID, key)
	if tokenErr!=nil{
		respondWithError(w, http.StatusBadRequest, tokenErr.Error())
	}
	type refreshResponseType struct{
		Token string `json:"token"`
	}
	refreshRes:=refreshResponseType{
		Token: accessToken,
	}
	respondWithJSON(w, http.StatusOK, refreshRes)
}

func (cfg *apiConfig) revoke(w http.ResponseWriter, r *http.Request){
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err!=nil{
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}
	validRefreshToken, validErr := cfg.dbQueries.GetRefreshToken(r.Context(),refreshToken)
	if validErr != nil{
		respondWithError(w, http.StatusUnauthorized, validErr.Error())
		return
	}
	_, revokeErr := cfg.dbQueries.RevokeRefreshToken(r.Context(),validRefreshToken.Token)
	if revokeErr!=nil{
		respondWithError(w, http.StatusUnauthorized, revokeErr.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) updateUser(w http.ResponseWriter, r *http.Request){
	type updateUserRequest struct{
		Email string `json:"email"`
		Password string `json:"password"`
	}
	userReq:=updateUserRequest{}
	err:=json.NewDecoder(r.Body).Decode(&userReq)
	if err!=nil{
		respondWithError(w, http.StatusBadRequest, "Failed to parse request")
		return
	}
	accessToken, tokenErr := auth.GetBearerToken(r.Header)
	if tokenErr!=nil{
		respondWithError(w, http.StatusUnauthorized, tokenErr.Error())
		return
	}
	key,_ :=os.LookupEnv("JWT_SECRET")
	userId, validErr := auth.ValidateJWT(accessToken,key)
	if validErr!=nil{
		respondWithError(w, http.StatusUnauthorized, validErr.Error())
		return
	}
	hashedPassword, hasErr := auth.HashPassword(userReq.Password)
	if hasErr!=nil{
		respondWithError(w, http.StatusBadRequest, hasErr.Error())
		return
	}
	userRes, updateErr := cfg.dbQueries.UpdateUser(r.Context(), database.UpdateUserParams{
		Email: userReq.Email,
		HashedPassword: hashedPassword,
		ID: userId,
	})
	if updateErr!=nil{
		respondWithError(w, http.StatusBadRequest, updateErr.Error())
		return
	}
	type userResponse struct{
		ID string `json:"id"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		Email string `json:"email"`
	}
	response := userResponse{
		ID: userRes.ID.String(),
		CreatedAt: userRes.CreatedAt.GoString(),
		UpdatedAt: userRes.UpdatedAt.GoString(),
		Email: userRes.Email,
	}
	respondWithJSON(w, http.StatusOK, response)
}

func (cfg *apiConfig) deleteChirp(w http.ResponseWriter, r *http.Request){
	chirpId := r.PathValue("chirpId")
	chirpUuid, parseError := uuid.Parse(chirpId)
	if parseError != nil{
		respondWithError(w, http.StatusBadRequest, "Could not parsed chirp id")
	}
	accessToken, tokenErr := auth.GetBearerToken(r.Header)
	if tokenErr!=nil{
		respondWithError(w, http.StatusUnauthorized, tokenErr.Error())
		return
	}
	key,_ :=os.LookupEnv("JWT_SECRET")
	userId, validErr := auth.ValidateJWT(accessToken,key)
	if validErr!=nil{
		respondWithError(w, http.StatusUnauthorized, validErr.Error())
		return
	}
	chirpData, chirpErr := cfg.dbQueries.GetChirpById(r.Context(),chirpUuid)
	if chirpErr != nil{
		respondWithError(w, http.StatusNotFound,chirpErr.Error())
		return
	}
	if chirpData.UserID != userId{
		respondWithError(w, http.StatusForbidden, "You are not the owner of this chirp")
		return
	}
	deleteErr := cfg.dbQueries.DeleteChirp(r.Context(), chirpUuid)
	if deleteErr!=nil{
		respondWithError(w, http.StatusBadRequest, deleteErr.Error())
		return
	}
	respondWithJSON(w, http.StatusNoContent, struct{}{})
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err!=nil{
		log.Println("Error connecting to database")
		log.Fatal("Error connecting to database")
	}
	dbQueries := database.New(db)

	apiMiddleware := apiConfig{
		dbQueries: dbQueries,
		platform: os.Getenv("PLATFORM"),
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.Handle(
		"/app/",
		apiMiddleware.middlewareMetricsInc(
			http.StripPrefix(
				"/app/",
				http.FileServer(http.Dir(".")),
			),
		),
	)

	mux.HandleFunc("GET /admin/metrics", apiMiddleware.metricsHandler)
	mux.HandleFunc("POST /admin/reset", apiMiddleware.resetHandler)
	mux.HandleFunc("POST /api/users", apiMiddleware.addUser)
	mux.HandleFunc("POST /api/chirps", apiMiddleware.validateChirp)
	mux.HandleFunc("GET /api/chirps", apiMiddleware.getChirps)
	mux.HandleFunc("GET /api/chirps/{chirpId}", apiMiddleware.getChirpById)
	mux.HandleFunc("POST /api/login", apiMiddleware.login)
	mux.HandleFunc("POST /api/refresh", apiMiddleware.refresh)
	mux.HandleFunc("POST /api/revoke", apiMiddleware.revoke)
	mux.HandleFunc("PUT /api/users", apiMiddleware.updateUser)
	mux.HandleFunc("DELETE /api/chirps/{chirpId}", apiMiddleware.deleteChirp)

	s := &http.Server{
		Addr:           ":8080",
		Handler:        middleware.MiddlewareLog(mux),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Println("Server starting on http://localhost:8080")

	log.Fatal(s.ListenAndServe())
}