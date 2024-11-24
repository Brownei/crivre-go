package api

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/brownei/crivre-go/types"
	"github.com/brownei/crivre-go/utils"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/markbates/goth/gothic"
)

func (a *application) AllAuthRoutes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(a.AuthMiddleware)
		// r.Get("/user", a.GetCurrentUser)
	})
	r.Post("/signin", a.Login)
	// r.Post("/signup", a.CreateAUser)
	r.Get("/{provider}", a.GoogleAuthLoginAndRegister)
	r.Get("/{provider}/callback", a.ProviderAuthCallbackFunction)
}

func (a *application) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	email := ctx.Value("user").(string)
	a.logger.Infof("Current user email: %s", email)

	existingUSer, err := a.store.User.GetUsersByEmail(ctx, email, true)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
	} else if existingUSer == nil {
		a.logger.Infof("Current user email: %v", existingUSer)
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("No user like this found"))
	}

	utils.WriteJSON(w, http.StatusOK, existingUSer)
}

func (a *application) GoogleAuthLoginAndRegister(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	r = r.WithContext(context.WithValue(context.Background(), "provider", provider))
	gothic.BeginAuthHandler(w, r)
}

func (a *application) CreateAUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var payload types.RegisterUserPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		log.Printf("Error: %s", err)
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	// Validate the payload
	if err := utils.Validator.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		fmt.Printf("Error: %s", errors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("Invalid payload: %v", errors))
		return
	}

	// Check if user exists first
	existingUser, err := a.store.User.GetUsersByEmail(ctx, payload.Email, false)
	log.Printf("existingUser: %v\n", existingUser)
	if existingUser != nil {
		utils.WriteError(w, http.StatusFound, fmt.Errorf("User already exists"))
		return
	}

	_, err = a.store.User.CreateNewUser(ctx, types.RegisterUserPayload{
		Email:          payload.Email,
		FirstName:      payload.FirstName,
		Password:       payload.Password,
		LastName:       payload.LastName,
		EmailVerified:  false,
		ProfilePicture: payload.ProfilePicture,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	token := utils.JwtToken(payload.Email, ctx)
	log.Printf("Token: %s", token)
	utils.WriteJSON(w, http.StatusCreated, token)
}

func (a *application) ProviderAuthCallbackFunction(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	r = r.WithContext(context.WithValue(context.Background(), "provider", provider))
	ctx := r.Context()

	if gothUser, err := gothic.CompleteUserAuth(w, r); err == nil {
		existingUSer, err := a.store.User.GetUsersByEmail(ctx, gothUser.Email, false)
		if err != nil {
			hashedPassword, err := utils.HashPassword(gothUser.Email)
			if err != nil {
				utils.WriteError(w, http.StatusBadRequest, err)
			}

			_, err = a.store.User.CreateNewUser(ctx, types.RegisterUserPayload{
				Email:          gothUser.Email,
				FirstName:      gothUser.FirstName,
				LastName:       gothUser.LastName,
				ProfilePicture: gothUser.AvatarURL,
				Password:       hashedPassword,
				EmailVerified:  true,
			})

			token := utils.JwtToken(gothUser.Email, ctx)
			response := fmt.Sprintf(`
        <script>
            window.opener.postMessage({ token: "%s" }, "*");
            window.close();
        </script>
    `, token)
			// a.logger.Info(token)
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(response))
		} else if existingUSer != nil {
			token := utils.JwtToken(existingUSer.Email, ctx)
			// utils.WriteJSON(w, http.StatusOK, token)

			response := fmt.Sprintf(`
        <script>
            window.opener.postMessage({ token: "%s" }, "*");
            window.close();
        </script>
    `, token)
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(response))

		}

	} else {
		utils.WriteError(w, http.StatusBadRequest, err)
	}
}

func (a *application) Login(w http.ResponseWriter, r *http.Request) {
	var loginPayload types.LoginPayload
	ctx := r.Context()
	if err := utils.ParseJSON(r, &loginPayload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
	}
	existingUser, err := a.store.User.GetUsersByEmail(ctx, loginPayload.Email, true)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
	} else if existingUser == nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("No user like this exists"))
	}

	token, err := a.store.Auth.Login(ctx, existingUser.Password, loginPayload.Password, existingUser.Email)
	if err != nil {
		utils.WriteError(w, http.StatusConflict, err)
	}

	utils.WriteJSON(w, http.StatusAccepted, token)
}
